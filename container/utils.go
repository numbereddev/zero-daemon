package container

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/buger/jsonparser"
	"github.com/moby/moby/client"
)

// checkImage checks an image and potentially pulls it, it's a non-blocking function and outputs
// the status updates like docker log lines or errors through the channel
func (c *Container) checkImage(image string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// TODO: for now it just always pulls but we gotta make this smarter

	c.bus.Publish(EventImagePullBegin, fmt.Sprintf("pulling image `%s`...", image))
	out, err := c.cli.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer out.Close()

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Bytes()

		status, _ := jsonparser.GetString(line, "status")
		current, _ := jsonparser.GetUint64(line, "progressDetail", "current")
		total, _ := jsonparser.GetUint64(line, "progressDetail", "total")

		if payload, err := json.Marshal(ImagePullProgress{
			Status:  status,
			Current: current,
			Total:   total,
		}); err != nil {
			return err
		} else {
			c.bus.Publish(EventImagePullProgress, string(payload))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	c.bus.Publish(EventImagePullDone, fmt.Sprintf("finished pulling image `%s`.", image))
	return nil
}
