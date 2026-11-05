package rest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/internal/version"
	"github.com/yestefir/micup.go/src/model"
)

type Upload struct {
	Name        string
	ContentType string
	Description string
	Size        int64
	Reader      io.Reader
}

type uploadPlan struct {
	Attachments []struct {
		UploadFilename string `json:"upload_filename"`
		ContentType    string `json:"content_type"`
		UploadMode     string `json:"upload_mode"`
		UploadURL      string `json:"upload_url"`
		UploadID       string `json:"upload_id"`
		PartSize       int64  `json:"part_size"`
		Parts          []struct {
			PartNumber int    `json:"part_number"`
			UploadURL  string `json:"upload_url"`
		} `json:"parts"`
	} `json:"attachments"`
}

func (c *Client) UploadAttachment(ctx context.Context, channelID string, upload Upload) (model.UploadedAttachment, error) {
	if channelID == "" {
		return model.UploadedAttachment{}, errors.New("channel ID is required")
	}
	if upload.Name == "" || upload.Reader == nil || upload.Size <= 0 {
		return model.UploadedAttachment{}, errors.New("an upload needs a name, a reader and its exact size")
	}
	contentType := contentTypeOr(upload.ContentType)
	var plan uploadPlan
	body := map[string]any{"attachments": []map[string]any{{"id": 0, "filename": upload.Name, "file_size": upload.Size, "content_type": contentType}}}
	if err := c.call(ctx, &request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "attachments"), body: body, dest: &plan}); err != nil {
		return model.UploadedAttachment{}, err
	}
	if len(plan.Attachments) != 1 {
		return model.UploadedAttachment{}, errors.New("Micup returned an invalid upload plan")
	}
	item := plan.Attachments[0]
	if item.ContentType != "" {
		contentType = item.ContentType
	}
	switch item.UploadMode {
	case "singlepart":
		if err := c.put(ctx, item.UploadURL, contentType, io.LimitReader(upload.Reader, upload.Size), upload.Size); err != nil {
			return model.UploadedAttachment{}, err
		}
	case "multipart":
		if item.PartSize <= 0 || len(item.Parts) == 0 {
			return model.UploadedAttachment{}, errors.New("Micup returned an invalid multipart upload plan")
		}
		remaining := upload.Size
		for _, part := range item.Parts {
			size := min(item.PartSize, remaining)
			if size <= 0 {
				return model.UploadedAttachment{}, errors.New("Micup upload plan has more parts than the file")
			}
			chunk := make([]byte, size)
			if _, err := io.ReadFull(upload.Reader, chunk); err != nil {
				return model.UploadedAttachment{}, fmt.Errorf("reading upload part %d: %w", part.PartNumber, err)
			}
			if err := c.put(ctx, part.UploadURL, "", bytes.NewReader(chunk), size); err != nil {
				return model.UploadedAttachment{}, err
			}
			remaining -= size
		}
		if remaining != 0 {
			return model.UploadedAttachment{}, errors.New("Micup upload plan does not cover the whole file")
		}
		complete := map[string]any{"uploads": []map[string]string{{"upload_filename": item.UploadFilename, "upload_id": item.UploadID}}}
		if err := c.call(ctx, &request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "attachments", "complete"), body: complete}); err != nil {
			return model.UploadedAttachment{}, err
		}
	default:
		return model.UploadedAttachment{}, fmt.Errorf("unsupported Micup upload mode %q", item.UploadMode)
	}
	return model.UploadedAttachment{Filename: upload.Name, UploadFilename: item.UploadFilename, Size: upload.Size, ContentType: contentType, Description: upload.Description}, nil
}

func (c *Client) put(ctx context.Context, target, contentType string, body io.Reader, size int64) error {
	parsed, err := url.Parse(target)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		return errors.New("Micup returned an invalid upload URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, target, body)
	if err != nil {
		return err
	}
	req.ContentLength = size
	req.Header.Set("User-Agent", version.UserAgent)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("Micup attachment upload failed")
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, discardLimit))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Micup attachment upload returned HTTP %d", resp.StatusCode)
	}
	return nil
}
