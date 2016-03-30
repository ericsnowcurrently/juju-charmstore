// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package v5

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"gopkg.in/errgo.v1"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/juju/charm.v6-unstable/resource"
	"gopkg.in/juju/charmrepo.v2-unstable/csclient/params"

	"gopkg.in/juju/charmstore.v5-unstable/internal/charmstore"
	"gopkg.in/juju/charmstore.v5-unstable/internal/mongodoc"
	"gopkg.in/juju/charmstore.v5-unstable/internal/router"
)

var metaResourceRE = regexp.MustCompile(`^/?([^/$]+)(?:/(\d+))?$`)

// GET id/meta/resource/name
// https://github.com/juju/charmstore/blob/v5/docs/API.md#get-idmetaresource
func (h *ReqHandler) metaResource(entity *mongodoc.Entity, id *router.ResolvedURL, path string, flags url.Values, req *http.Request) (interface{}, error) {
	if entity.URL.Series == "bundle" {
		// Bundles do not have resources so we return an empty result.
		return params.Resource{}, nil
	}
	if entity.CharmMeta == nil {
		// This shouldn't happen...
		panic("entity missing charm metadata")
	}

	parts := metaResourceRE.FindStringSubmatch(path)
	if parts == nil {
		return params.Resource{}, errgo.WithCausef(nil, params.ErrBadRequest, "bad path for resource info")
	}
	resName := parts[1]
	revStr := parts[2]

	// TODO(ericsnow) Handle flags.
	var doc *mongodoc.Resource
	if revStr == "" {
		// TODO(ericsnow) Add Store.LatestResourceInfo()?
		docs, err := h.Store.ListResources(entity, params.UnpublishedChannel)
		if err != nil {
			return params.Resource{}, errgo.Mask(err)
		}
		doc = &mongodoc.Resource{Name: resName} // used if not found
		for _, actual := range docs {
			if actual.Name == resName {
				doc = actual
			}
		}
		// TODO(ericsnow) Fail if not found?
	} else {
		revision, _ := strconv.Atoi(revStr) // The regex guarantees an int.
		actual, err := h.Store.ResourceInfo(entity, resName, revision)
		if charmstore.IsResourceNotFound(err) {
			// TODO(ericsnow) Fail?
			actual = &mongodoc.Resource{Name: resName}
		} else if err != nil {
			return params.Resource{}, errgo.Mask(err)
		}
		doc = actual
	}
	result := Resource2API(doc, entity.CharmMeta)
	return result, nil
}

// GET id/meta/resources
// https://github.com/juju/charmstore/blob/v5/docs/API.md#get-idmetaresources
func (h *ReqHandler) metaResources(entity *mongodoc.Entity, id *router.ResolvedURL, path string, flags url.Values, req *http.Request) (interface{}, error) {
	if entity.URL.Series == "bundle" {
		// Bundles do not have resources so we return an empty result.
		return []params.Resource{}, nil
	}
	if entity.CharmMeta == nil {
		// This shouldn't happen...
		panic("entity missing charm metadata")
	}

	channel, err := h.entityChannel(id)
	if err != nil {
		// Given the current implementation of entityChannel(), this
		// should not fail. However, we handle the error here for the
		// sake of future changes to entityChannel().
		return nil, errgo.Mask(err)
	}

	// TODO(ericsnow) Handle flags.
	docs, err := h.Store.ListResources(entity, channel)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	var results []params.Resource
	for _, doc := range docs {
		result := Resource2API(doc, entity.CharmMeta)
		results = append(results, result)
	}
	return results, nil
}

// POST id/resources/name
// https://github.com/juju/charmstore/blob/v5/docs/API.md#post-idresourcesname
//
// GET  id/resources/name[/revision]
// https://github.com/juju/charmstore/blob/v5/docs/API.md#get-idresourcesnamerevision
func (h *ReqHandler) serveResources(id *router.ResolvedURL, w http.ResponseWriter, req *http.Request) error {
	// Resources are "published" using "PUT id/publish" so we don't
	// support PUT here.
	// TODO(ericsnow) Support DELETE to remove a resource?
	// (like serveArchive() does)
	switch req.Method {
	case "GET":
		return h.serveDownloadResource(id, w, req)
	case "POST":
		return h.serveUploadResource(id, w, req)
	default:
		return errgo.WithCausef(nil, params.ErrMethodNotAllowed, "%s not allowed", req.Method)
	}
}

func (h *ReqHandler) serveDownloadResource(id *router.ResolvedURL, w http.ResponseWriter, req *http.Request) error {
	return errNotImplemented
}

func (h *ReqHandler) serveUploadResource(id *router.ResolvedURL, w http.ResponseWriter, req *http.Request) error {
	return errNotImplemented
}

// Resource2API converts a resource doc into the corresponding
// params type.
func Resource2API(doc *mongodoc.Resource, chMeta *charm.Meta) params.Resource {
	apiRes := resource2api(doc, chMeta)
	if len(doc.Fingerprint) == 0 {
		// We ensure that the fingerprint isn't nil in order
		// to produce consistent results.
		apiRes.Fingerprint = []byte{}
	}
	return apiRes
}

func resource2api(doc *mongodoc.Resource, chMeta *charm.Meta) params.Resource {
	meta, inMeta := chMeta.Resources[doc.Name]
	if !inMeta {
		return params.Resource{}
	}

	apiRes := params.Resource{
		Name:        meta.Name,
		Type:        meta.Type.String(),
		Path:        meta.Path,
		Description: meta.Description,
		Origin:      resource.OriginStore.String(),
		Revision:    doc.Revision,
		Fingerprint: doc.Fingerprint,
		Size:        doc.Size,
	}
	if len(doc.Fingerprint) == 0 {
		// The resource has not been uploaded yet. Hence it must be
		// provided directly by the user to the Juju controller. We
		// indicate this by changing the origin to "upload".
		apiRes.Origin = resource.OriginUpload.String()
	}
	return apiRes
}
