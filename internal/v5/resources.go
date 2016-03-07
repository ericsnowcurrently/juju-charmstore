// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package v5 // import "gopkg.in/juju/charmstore.v5-unstable/internal/v5"

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"gopkg.in/errgo.v1"
	"gopkg.in/juju/charm.v6-unstable/resource"
	"gopkg.in/juju/charmrepo.v2-unstable/csclient/params"

	"gopkg.in/juju/charmstore.v5-unstable/internal/mongodoc"
	"gopkg.in/juju/charmstore.v5-unstable/internal/router"
)

// GET id/meta/resources
// https://github.com/juju/charmstore/blob/v5/docs/API.md#get-idmetaresources
func (h *ReqHandler) metaResources(entity *mongodoc.Entity, id *router.ResolvedURL, path string, flags url.Values, req *http.Request) (interface{}, error) {
	return h.Store.ListResources(entity)
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
	_, err := h.AuthorizeEntityAndTerms(req, []*router.ResolvedURL{id})
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}
	name, revision, err := extractResourceRequest(req)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}
	res, reader, err := h.Store.OpenResourceBlob(id, name, revision)
	if err != nil {
		return errgo.Mask(err, errgo.Is(params.ErrNotFound))
	}
	defer reader.Close()
	h.sendResource(id, res, w, req, reader)
	return nil
}

func extractResourceRequest(req *http.Request) (name string, revision int, err error) {
	reqPath := strings.TrimPrefix(path.Clean(req.URL.path), "/")
	parts := strings.SplitN(reqPath, "/", 2)
	name = parts[0]
	revision = -1 // "use the latest"
	if len(parts) == 2 {
		rev, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", 0, errgo.Notef(err, "invalid resource revision in URL")
		}
		revision = rev
	}
	return name, revision, nil
}

func (h *ReqHandler) sendResource(id *router.ResolvedURL, res resource.Resource, w http.ResponseWriter, req *http.Request, reader io.ReadCloser) {
	header := w.Header()
	// TODO(ericsnow) Use a separate resource cache control?
	setArchiveCacheControl(w.Header(), h.isPublic(id))
	hash := res.Fingerprint.String()
	logger.Infof("sendResource setting %s=%s", params.ContentHashHeader, hash)
	header.Set(params.ContentHashHeader, hash)
	header.Set(params.EntityIdHeader, id.PreferredURL().String())
	resID := fmt.Sprintf("%s/%d", res.Name, res.Revision)
	header.Set(params.ResourceIdHeader, resID)

	// TODO(ericsnow) track stats for resource downloads?

	// TODO(rog) should we set connection=close here?
	// See https://codereview.appspot.com/5958045
	timestamp := time.Time{}
	http.ServeContent(w, req, res.Path, timestamp, reader)
}

func (h *ReqHandler) serveUploadResource(id *router.ResolvedURL, w http.ResponseWriter, req *http.Request) error {
	return errNotImplemented
}
