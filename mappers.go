package crud

import (
	"context"
	"fmt"
)

// Relation defines the behavior for an eager-loading process for a given parent type T.
// This interface allows different relationship types (many-to-one, one-to-many, etc.)
// to be handled polymorphically.
type Relation[T any] interface {
	// Process executes the logic for a given relationship.
	// It fetches the related entities and maps them back to the parents.
	Process(ctx context.Context, parents []*T) error
}

// --- ManyToOneMapper ---

// ManyToOneMapper implements the Relation interface for a many-to-one (Belongs To) relationship.
// ParentT is the type of the model being queried (e.g., Post).
// RelatedT is the type of the model to be loaded (e.g., User).
// FKT is the type of the foreign key (e.g., int).
type ManyToOneMapper[ParentT any, RelatedT any, FKT comparable] struct {
	// Fetcher is the function that retrieves the related models by their keys.
	Fetcher RelatedFetcher[FKT, RelatedT]
	// GetFK extracts the foreign key from the parent model.
	GetFK func(p *ParentT) FKT
	// GetPK extracts the primary key from the related model.
	GetPK func(r *RelatedT) FKT
	// SetRelated sets the single related model onto the parent model.
	SetRelated func(p *ParentT, r *RelatedT)
}

// Process executes the eager loading logic for the many-to-one relationship.
func (m ManyToOneMapper[ParentT, RelatedT, FKT]) Process(ctx context.Context, parents []*ParentT) error {
	if m.Fetcher == nil || m.GetFK == nil || m.GetPK == nil || m.SetRelated == nil {
		return fmt.Errorf("ManyToOneMapper is not fully configured")
	}

	keyMap := make(map[FKT]struct{})
	var keys []FKT
	for _, p := range parents {
		fk := m.GetFK(p)
		// Add key to the list if it's not zero and not already present
		var zero FKT
		if fk != zero {
			if _, exists := keyMap[fk]; !exists {
				keyMap[fk] = struct{}{}
				keys = append(keys, fk)
			}
		}
	}

	if len(keys) == 0 {
		return nil
	}

	related, err := m.Fetcher(ctx, keys)
	if err != nil {
		return fmt.Errorf("failed to fetch related entities for ManyToOne: %w", err)
	}

	relatedMap := make(map[FKT]RelatedT)
	for i := range related {
		rel := &related[i]
		pk := m.GetPK(rel)
		relatedMap[pk] = *rel
	}

	for _, p := range parents {
		fk := m.GetFK(p)
		if rel, found := relatedMap[fk]; found {
			m.SetRelated(p, &rel)
		}
	}
	return nil
}

// --- OneToManyMapper ---

// OneToManyMapper implements the Relation interface for a one-to-many (Has Many) relationship.
type OneToManyMapper[ParentT any, RelatedT any, PKT comparable] struct {
	Fetcher    RelatedFetcher[PKT, RelatedT]
	GetPK      func(p *ParentT) PKT
	GetFK      func(r *RelatedT) PKT
	SetRelated func(p *ParentT, r []*RelatedT)
}

// Process executes the eager loading logic for the one-to-many relationship.
func (m OneToManyMapper[ParentT, RelatedT, PKT]) Process(ctx context.Context, parents []*ParentT) error {
	if m.Fetcher == nil || m.GetPK == nil || m.GetFK == nil || m.SetRelated == nil {
		return fmt.Errorf("OneToManyMapper is not fully configured")
	}

	var keys []PKT
	for _, p := range parents {
		keys = append(keys, m.GetPK(p))
	}

	if len(keys) == 0 {
		return nil
	}

	related, err := m.Fetcher(ctx, keys)
	if err != nil {
		return fmt.Errorf("failed to fetch related entities for OneToMany: %w", err)
	}

	groupedRelated := make(map[PKT][]*RelatedT)
	for i := range related {
		rel := &related[i]
		fk := m.GetFK(rel)
		groupedRelated[fk] = append(groupedRelated[fk], rel)
	}

	for _, p := range parents {
		pk := m.GetPK(p)
		// Always set the slice, even if it's empty, to distinguish between nil (not loaded) and empty (no related items).
		if rels, found := groupedRelated[pk]; found {
			m.SetRelated(p, rels)
		} else {
			m.SetRelated(p, make([]*RelatedT, 0))
		}
	}
	return nil
}

// --- HasOneMapper ---

// HasOneMapper implements the Relation interface for a one-to-one (Has One) relationship.
type HasOneMapper[ParentT any, RelatedT any, PKT comparable] struct {
	Fetcher    RelatedFetcher[PKT, RelatedT]
	GetPK      func(p *ParentT) PKT
	GetFK      func(r *RelatedT) PKT
	SetRelated func(p *ParentT, r *RelatedT)
}

// Process executes the eager loading logic for the one-to-one relationship.
func (m HasOneMapper[ParentT, RelatedT, PKT]) Process(ctx context.Context, parents []*ParentT) error {
	if m.Fetcher == nil || m.GetPK == nil || m.GetFK == nil || m.SetRelated == nil {
		return fmt.Errorf("HasOneMapper is not fully configured")
	}

	var keys []PKT
	for _, p := range parents {
		keys = append(keys, m.GetPK(p))
	}

	if len(keys) == 0 {
		return nil
	}

	related, err := m.Fetcher(ctx, keys)
	if err != nil {
		return fmt.Errorf("failed to fetch related entities for HasOne: %w", err)
	}

	relatedMap := make(map[PKT]RelatedT)
	for i := range related {
		rel := &related[i]
		fk := m.GetFK(rel)
		relatedMap[fk] = *rel
	}

	for _, p := range parents {
		pk := m.GetPK(p)
		if rel, found := relatedMap[pk]; found {
			m.SetRelated(p, &rel)
		}
	}
	return nil
}
