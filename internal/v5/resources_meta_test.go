// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package v5_test

import (
	gc "gopkg.in/check.v1"
)

type ResourcesMetaSuite struct {
	commonSuite
}

var _ = gc.Suite(&ResourcesMetaSuite{})

func (s *ResourcesMetaSuite) TestResourceWithRevisionFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceWithRevisionNotFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceWithoutRevisionFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceWithoutRevisionNotFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceBundle(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceBadPath(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourcesPublishedFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourcesNotPublishedWithChannel(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourcesNotPublishedWithoutChannel(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourcesNotPublishedNotFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourcesBundle(c *gc.C) {
}
