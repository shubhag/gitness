// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package space

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	apiauth "github.com/harness/gitness/internal/api/auth"
	"github.com/harness/gitness/internal/api/usererror"
	"github.com/harness/gitness/internal/auth"
	"github.com/harness/gitness/internal/paths"
	"github.com/harness/gitness/internal/store"
	"github.com/harness/gitness/internal/store/database/dbtx"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/check"
	"github.com/harness/gitness/types/enum"
)

// MoveInput is used for moving a space.
type MoveInput struct {
	UID         *string `json:"uid"`
	ParentID    *int64  `json:"parent_id"`
	KeepAsAlias bool    `json:"keep_as_alias"`
}

func (i *MoveInput) hasChanges(space *types.Space) bool {
	return (i.UID != nil && *i.UID != space.UID) ||
		(i.ParentID != nil && *i.ParentID != space.ParentID)
}

// Move moves a space to a new space and/or name.
//
//nolint:gocognit // refactor if needed
func (c *Controller) Move(ctx context.Context, session *auth.Session,
	spaceRef string, in *MoveInput) (*types.Space, error) {
	space, err := c.spaceStore.FindByRef(ctx, spaceRef)
	if err != nil {
		return nil, err
	}

	permission := enum.PermissionSpaceEdit
	if in.ParentID != nil && *in.ParentID != space.ParentID {
		// ensure user has access to new space (parentId not sanitized!)
		if err = c.checkAuthSpaceCreation(ctx, session, *in.ParentID); err != nil {
			return nil, fmt.Errorf("failed to verify space creation permissions on new parent space: %w", err)
		}

		// TODO: what would be correct permissions on space? (technically we are deleting it from the old space)
		permission = enum.PermissionSpaceDelete
	}
	if err = apiauth.CheckSpace(ctx, c.authorizer, session, space, permission, false); err != nil {
		return nil, err
	}

	if !in.hasChanges(space) {
		return space, nil
	}

	if err = c.sanitizeMoveInput(in, space.ParentID == 0); err != nil {
		return nil, fmt.Errorf("failed to sanitize input: %w", err)
	}

	err = dbtx.New(c.db).WithTx(ctx, func(ctx context.Context) error {
		space, err = c.spaceStore.UpdateOptLock(ctx, space, func(s *types.Space) error {
			if in.UID != nil {
				s.UID = *in.UID
			}
			if in.ParentID != nil {
				s.ParentID = *in.ParentID
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to update space: %w", err)
		}

		// lock path to ensure it doesn't get updated while we move the space (and its descendants)
		var primaryPath *types.Path
		primaryPath, err = c.pathStore.FindPrimaryWithLock(ctx, enum.PathTargetTypeSpace, space.ID)
		if err != nil {
			return fmt.Errorf("failed to find primary path: %w", err)
		}

		oldPathValue := primaryPath.Value
		newPathValue := space.UID
		if space.ParentID > 0 {
			var parentPath *types.Path
			parentPath, err = c.pathStore.FindPrimary(ctx, enum.PathTargetTypeSpace, space.ParentID)
			if err != nil {
				return fmt.Errorf("failed to find parent space path: %w", err)
			}
			newPathValue = paths.Concatinate(parentPath.Value, space.UID)
		}
		space.Path = newPathValue

		// ensure we don't move space into itself
		if strings.HasPrefix(newPathValue, oldPathValue+types.PathSeparator) {
			return usererror.ErrCyclicHierarchy
		}

		var descendantPaths []*types.Path
		descendantPaths, err = c.pathStore.ListPrimaryDescendantsWithLock(ctx, oldPathValue)
		if err != nil {
			return fmt.Errorf("failed to list all primary descendants: %w", err)
		}

		return c.movePaths(ctx, session.Principal.ID, append(descendantPaths, primaryPath),
			oldPathValue, newPathValue, in.KeepAsAlias)
	})
	if err != nil {
		return nil, err
	}

	return space, nil
}

func (c *Controller) sanitizeMoveInput(in *MoveInput, isRoot bool) error {
	if in.ParentID != nil {
		if *in.ParentID < 0 {
			return errParentIDNegative
		}
		isRoot = *in.ParentID == 0
	}

	if in.UID != nil {
		if err := c.uidCheck(*in.UID, isRoot); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) movePaths(ctx context.Context, principalID int64, paths []*types.Path,
	oldPathPrefix string, newPathPrefix string, keepAsAlias bool) error {
	for _, p := range paths {
		oldValue := p.Value
		p.Value = newPathPrefix + p.Value[len(oldPathPrefix):]

		err := check.PathDepth(p.Value, p.TargetType == enum.PathTargetTypeSpace)
		if err != nil {
			return usererror.BadRequestf("resulting path '%s' failed validation: %s", p.Value, err)
		}

		err = c.pathStore.Update(ctx, p)
		if errors.Is(err, store.ErrDuplicate) {
			return usererror.BadRequestf("resulting path '%s' already exists", p.Value)
		}
		if err != nil {
			return fmt.Errorf("failed to update primary path for '%s': %w", oldValue, err)
		}

		if keepAsAlias {
			now := time.Now().UnixMilli()
			err = c.pathStore.Create(ctx, &types.Path{
				Version:    0,
				Value:      oldValue,
				IsPrimary:  false,
				TargetType: p.TargetType,
				TargetID:   p.TargetID,
				CreatedBy:  principalID,
				Created:    now,
				Updated:    now,
			})
			if err != nil {
				return fmt.Errorf("failed to create alias path '%s': %w", oldValue, err)
			}
		}
	}

	return nil
}