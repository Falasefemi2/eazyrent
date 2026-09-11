package listing

import "errors"

// ErrListingNotFound maps a missing listing to 404 with errors.Is.
// ErrListingForbidden maps a non-owner mutation to 403.
var (
	ErrListingNotFound  = errors.New("listing not found")
	ErrListingForbidden = errors.New("you don't own this listing")
)
