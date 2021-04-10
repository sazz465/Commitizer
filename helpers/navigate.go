package helpers

import (
	"context"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/pkg/errors"
)

// Navigate to the URL and waits for DOMContentEventFired. An error is
// returned if timeout happens before DOMContentEventFired.
func Navigate(ctx context.Context, pageClient cdp.Page, url string, domLoadTimeout time.Duration) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, domLoadTimeout)
	defer cancel()

	// Make sure Page events are enabled.
	err := pageClient.Enable(ctx)
	if err != nil {
		return errors.Wrap(err, "page events enabling failed")
	}

	// Open client for DOMContentEventFired to block until DOM has fully loaded.
	domContentEventFired, err := pageClient.DOMContentEventFired(ctx)
	if err != nil {
		return errors.Wrap(err, "For event DOMContentEventFired, blocking before DOM loading failed")
	}
	defer domContentEventFired.Close()

	_, err = pageClient.Navigate(ctx, page.NewNavigateArgs(url))
	if err != nil {
		return errors.Wrap(err, "couldn't navigate to page")
	}

	_, err = domContentEventFired.Recv()
	return errors.Wrap(err, "domContentEventFired event not received")
}
