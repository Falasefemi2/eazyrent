package listing

import (
	"context"

	"github.com/google/uuid"
)

// CreateListingParams carries the fields to create a listing. Nine fields
// travel together from the HTTP boundary, so one struct beats nine
// positional arguments.
type CreateListingParams struct {
	LandlordID  uuid.UUID
	Title       string
	Description string
	Price       string
	Rooms       int
	Furnished   bool
	Latitude    float64
	Longitude   float64
	Address     string
}

// Service enforces listing ownership over a Store, mirroring the TS
// ListingService minus its Redis cache: this project has no Redis, so
// reads always hit Postgres. Handlers call it directly.
type Service struct {
	Listings Store
}

func NewService(store Store) Service {
	return Service{Listings: store}
}

func assertOwner(l Listing, landlordID uuid.UUID) error {
	if l.LandlordID != landlordID {
		return ErrListingForbidden
	}
	return nil
}

// Create inserts one listing. Arguments are already validated by the caller.
func (s Service) Create(ctx context.Context, p CreateListingParams) (Listing, error) {
	return s.Listings.Create(ctx, p.LandlordID, p.Title, p.Description, p.Price,
		p.Rooms, p.Furnished, p.Latitude, p.Longitude, p.Address)
}

// GetByID returns a listing with media and landlord contact, or
// ErrListingNotFound.
func (s Service) GetByID(ctx context.Context, id uuid.UUID) (Detail, error) {
	return s.Listings.FindDetail(ctx, id)
}

// GetAll returns a page of filtered listings, newest first.
func (s Service) GetAll(ctx context.Context, page, limit int, f Filters) (Page, error) {
	return s.Listings.FindAll(ctx, page, limit, f)
}

// GetMyListings returns a page of one landlord's listings, newest first.
func (s Service) GetMyListings(ctx context.Context, landlordID uuid.UUID, page, limit int) (Page, error) {
	return s.Listings.FindByLandlord(ctx, landlordID, page, limit)
}

// AddMedia attaches an already-uploaded file URL to a listing. Uploads go
// straight to the media provider (see docs/avatar-uploads.md for the
// pattern); only the URL is stored. The caller must own the listing.
func (s Service) AddMedia(ctx context.Context, listingID, landlordID uuid.UUID, url, mediaType string, order int) (Media, error) {
	l, err := s.Listings.FindByID(ctx, listingID)
	if err != nil {
		return Media{}, err
	}
	if err := assertOwner(l, landlordID); err != nil {
		return Media{}, err
	}
	return s.Listings.AddMedia(ctx, listingID, url, mediaType, order)
}

// DeleteMedia removes one media row. Like the TS service it takes the
// media ID directly; the HTTP endpoint calling it must scope the media to
// a listing the caller owns before delegating here.
func (s Service) DeleteMedia(ctx context.Context, mediaID uuid.UUID) error {
	return s.Listings.DeleteMedia(ctx, mediaID)
}

// Update changes a landlord's listing, or returns ErrListingNotFound /
// ErrListingForbidden.
func (s Service) Update(ctx context.Context, id, landlordID uuid.UUID, p UpdateParams) (Listing, error) {
	l, err := s.Listings.FindByID(ctx, id)
	if err != nil {
		return Listing{}, err
	}
	if err := assertOwner(l, landlordID); err != nil {
		return Listing{}, err
	}
	return s.Listings.Update(ctx, id, p)
}

// UpdateStatus is Update for the status field alone.
func (s Service) UpdateStatus(ctx context.Context, id, landlordID uuid.UUID, status string) (Listing, error) {
	return s.Update(ctx, id, landlordID, UpdateParams{Status: &status})
}

// Delete removes a landlord's listing; media and favorites cascade.
func (s Service) Delete(ctx context.Context, id, landlordID uuid.UUID) error {
	l, err := s.Listings.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := assertOwner(l, landlordID); err != nil {
		return err
	}
	return s.Listings.Delete(ctx, id)
}
