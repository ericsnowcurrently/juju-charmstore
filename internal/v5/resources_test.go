// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package v5_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/juju/charm.v6-unstable/resource"
	"gopkg.in/juju/charmrepo.v2-unstable/csclient/params"

	"gopkg.in/juju/charmstore.v5-unstable/internal/charmstore"
	"gopkg.in/juju/charmstore.v5-unstable/internal/mongodoc"
	"gopkg.in/juju/charmstore.v5-unstable/internal/router"
	"gopkg.in/juju/charmstore.v5-unstable/internal/storetesting"
)

func addCharm(c *gc.C, store *charmstore.Store, curl *charm.URL) (*router.ResolvedURL, *mongodoc.Entity, *charm.CharmDir) {
	resolvedURL := newResolvedURL(curl.String(), curl.Revision)
	ch := storetesting.Charms.CharmDir(curl.Name)
	err := store.AddCharmWithArchive(resolvedURL, ch)
	c.Assert(err, jc.ErrorIsNil)
	entity, err := store.FindEntity(resolvedURL, nil)
	c.Assert(err, jc.ErrorIsNil)
	return resolvedURL, entity, ch
}

func addResources(c *gc.C, store *charmstore.Store, entity *mongodoc.Entity, ch *charm.CharmDir) (map[string]int, map[string]*bytes.Reader) {
	readers := extractResources(c, ch)
	revisions := make(map[string]int)
	for name, reader := range readers {
		revisions[name] = addResource(c, store, entity, name, reader)
		_, err := reader.Seek(0, os.SEEK_SET)
		c.Assert(err, jc.ErrorIsNil)
	}
	return revisions, readers
}

func addResource(c *gc.C, store *charmstore.Store, entity *mongodoc.Entity, resName string, blobReader io.ReadSeeker) int {
	blob := resourceBlob(c, blobReader)
	revision, err := store.AddResource(entity, resName, blob)
	c.Assert(err, jc.ErrorIsNil)
	_, err = blobReader.Seek(0, os.SEEK_SET)
	c.Assert(err, jc.ErrorIsNil)
	return revision
}

func extractResources(c *gc.C, ch *charm.CharmDir) map[string]*bytes.Reader {
	readers := make(map[string]*bytes.Reader)
	for _, meta := range ch.Meta().Resources {
		data, err := ioutil.ReadFile(filepath.Join(ch.Path, meta.Path))
		c.Assert(err, jc.ErrorIsNil)
		readers[meta.Name] = bytes.NewReader(data)
	}
	return readers
}

func resourceBlob(c *gc.C, r io.ReadSeeker) charmstore.ResourceBlob {
	var sizer utils.SizeTracker
	fp, err := resource.GenerateFingerprint(io.TeeReader(r, &sizer))
	c.Assert(err, jc.ErrorIsNil)
	_, err = r.Seek(0, os.SEEK_SET)
	c.Assert(err, jc.ErrorIsNil)
	return charmstore.ResourceBlob{
		Reader:      r,
		Fingerprint: fp.Bytes(),
		Size:        sizer.Size(),
	}
}

func publishResources(c *gc.C, store *charmstore.Store, entity *mongodoc.Entity, channel params.Channel, revisions map[string]int) {
	c.Assert(channel, gc.Not(gc.Equals), params.UnpublishedChannel)
	c.Assert(channel, gc.Not(gc.Equals), params.NoChannel)
	c.Assert(entity.CharmMeta.Resources, gc.HasLen, revisions)
	for name, revision := range revisions {
		_, ok := entity.CharmMeta.Resources[name]
		c.Assert(ok, jc.IsTrue)
		err := store.SetResource(entity, channel, name, revision)
		c.Assert(err, jc.ErrorIsNil)
	}
}
