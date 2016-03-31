// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package v5_test

import (
	"fmt"
	"net/http"
	"strings"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/testing/httptesting"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/juju/charmrepo.v2-unstable/csclient/params"

	"gopkg.in/juju/charmstore.v5-unstable/internal/v5"
)

type ResourcesMetaSuite struct {
	commonSuite
}

var _ = gc.Suite(&ResourcesMetaSuite{})

func (s *ResourcesMetaSuite) TestResourceWithRevisionFound(c *gc.C) {
	curl := charm.MustParseURL("cs:~charmers/utopic/starsay-17")
	id, entity, ch := addCharm(c, s.store, curl)
	s.setPublic(c, id)
	revisions, _ := addResources(c, s.store, entity, ch)
	blobReader := strings.NewReader("new data for for-store")
	revision := addResource(c, s.store, entity, "for-store", blobReader)
	c.Assert(revision, gc.Not(gc.Equals), revisions["for-store"])
	doc, err := s.store.ResourceInfo(entity, "for-store", revision)
	c.Assert(err, jc.ErrorIsNil)
	expected := v5.Resource2API(doc, entity.CharmMeta)

	s.checkResource(c, curl, "for-store", revision, expected)
}

func (s *ResourcesMetaSuite) TestResourceWithRevisionResourceNotFound(c *gc.C) {
	curl := charm.MustParseURL("cs:~charmers/utopic/starsay-17")
	id, entity, ch := addCharm(c, s.store, curl)
	s.setPublic(c, id)
	addResources(c, s.store, entity, ch)

	s.checkResource(c, curl, "who-dat", 0, nil)
}

func (s *ResourcesMetaSuite) TestResourceWithRevisionRevisionNotFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceWithoutRevisionFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceWithoutRevisionNotFound(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceBundle(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceBadPath(c *gc.C) {
}

func (s *ResourcesMetaSuite) TestResourceNotAuthorized(c *gc.C) {
	curl := charm.MustParseURL("cs:~charmers/utopic/starsay-17")
	_, entity, ch := addCharm(c, s.store, curl)
	revisions, _ := addResources(c, s.store, entity, ch)
	revision := revisions["for-store"]
	charmID := strings.TrimPrefix(curl.String(), "cs:")
	path := fmt.Sprintf("%s/meta/resource/for-store/%d", charmID, revision)

	s.assertGetIsUnauthorized(c, path, "authentication failed: missing HTTP auth header")
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

func (s *ResourcesMetaSuite) TestResourcesNotAuthorized(c *gc.C) {
}

func (s *ResourcesMetaSuite) checkResource(c *gc.C, curl *charm.URL, name string, revision int, expected interface{}) {
	charmID := strings.TrimPrefix(curl.String(), "cs:")
	path := fmt.Sprintf("%s/meta/resource/%s", charmID, name)
	if revision >= 0 {
		path += "/" + fmt.Sprint(revision)
	}

	if isNull(expected) {
		httptesting.AssertJSONCall(c, httptesting.JSONCallParams{
			Handler:      s.srv,
			URL:          storeURL(path),
			ExpectStatus: http.StatusNotFound,
			ExpectBody: params.Error{
				Message: params.ErrMetadataNotFound.Error(),
				Code:    params.ErrMetadataNotFound,
			},
		})
	} else {
		s.assertGet(c, path, expected)
	}
}
