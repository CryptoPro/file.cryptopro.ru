package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/expiry"
	"github.com/dustin/go-humanize"
	"github.com/flosch/pongo2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"github.com/zenazn/goji/web"
)

const maxDisplayFileSizeBytes = 1024 * 512

func fileDisplayHandler(c web.C, w http.ResponseWriter, r *http.Request, fileName string, metadata backends.Metadata) {
	if cpro {
		randomKey := fileName[len(fileName)-32:]
		fileName = fileName[:len(fileName)-33]
		directURL := "d/" + randomKey + "/" + fileName

		tpl := Templates["display/file.html"]
		err := renderTemplate(tpl, pongo2.Context{
			"mime":     metadata.Mimetype,
			"filename": fileName,
			"download": directURL,
			"size":     humanize.Bytes(uint64(metadata.Size)),
			"expiry": humanize.CustomRelTime(time.Now(), metadata.Expiry, "", "", []humanize.RelTimeMagnitude{
				{time.Second, "now", time.Second},
				{2 * time.Second, "1 second %s", 1},
				{time.Minute, "%d seconds %s", time.Second},
				{2 * time.Minute, "1 minute %s", 1},
				{time.Hour, "%d minutes %s", time.Minute},
				{2 * time.Hour, "1 hour %s", 1},
				{time.Hour * 96, "%d hours %s", time.Hour},
				{time.Hour * 90000, "%d days %s", time.Hour * 24},
			}),
			"expirylist":  listExpirationTimes(),
			"extra":       make(map[string]string),
			"forcerandom": Config.forceRandomFilename,
			"lines":       []string{},
			"files":       metadata.ArchiveFiles,
			"siteurl":     strings.TrimSuffix(getSiteURL(r), "/"),
		}, r, w)

		if err != nil {
			oopsHandler(c, w, r, RespHTML, "")
		}
		return
	}

	var expiryHuman string
	if metadata.Expiry != expiry.NeverExpire {
		expiryHuman = humanize.RelTime(time.Now(), metadata.Expiry, "", "")
	}
	sizeHuman := humanize.Bytes(uint64(metadata.Size))
	extra := make(map[string]string)
	lines := []string{}

	extension := strings.TrimPrefix(filepath.Ext(fileName), ".")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js, _ := json.Marshal(map[string]string{
			"filename":   fileName,
			"direct_url": getSiteURL(r) + Config.selifPath + fileName,
			"expiry":     strconv.FormatInt(metadata.Expiry.Unix(), 10),
			"size":       strconv.FormatInt(metadata.Size, 10),
			"mimetype":   metadata.Mimetype,
			"sha256sum":  metadata.Sha256sum,
		})
		w.Write(js)
		return
	}

	var tpl *pongo2.Template

	if strings.HasPrefix(metadata.Mimetype, "image/") {
		tpl = Templates["display/image.html"]

	} else if strings.HasPrefix(metadata.Mimetype, "video/") {
		tpl = Templates["display/video.html"]

	} else if strings.HasPrefix(metadata.Mimetype, "audio/") {
		tpl = Templates["display/audio.html"]

	} else if metadata.Mimetype == "application/pdf" {
		tpl = Templates["display/pdf.html"]

	} else if extension == "story" {
		metadata, reader, err := storageBackend.Get(fileName)
		if err != nil {
			oopsHandler(c, w, r, RespHTML, err.Error())
		}

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadAll(reader)
			if err == nil {
				extra["contents"] = string(bytes)
				lines = strings.Split(extra["contents"], "\n")
				tpl = Templates["display/story.html"]
			}
		}

	} else if extension == "md" {
		metadata, reader, err := storageBackend.Get(fileName)
		if err != nil {
			oopsHandler(c, w, r, RespHTML, err.Error())
		}

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadAll(reader)
			if err == nil {
				unsafe := blackfriday.MarkdownCommon(bytes)
				html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

				extra["contents"] = string(html)
				tpl = Templates["display/md.html"]
			}
		}

	} else if strings.HasPrefix(metadata.Mimetype, "text/") || supportedBinExtension(extension) {
		metadata, reader, err := storageBackend.Get(fileName)
		if err != nil {
			oopsHandler(c, w, r, RespHTML, err.Error())
		}

		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadAll(reader)
			if err == nil {
				extra["extension"] = extension
				extra["lang_hl"] = extensionToHlLang(extension)
				extra["contents"] = string(bytes)
				tpl = Templates["display/bin.html"]
			}
		}
	}

	// Catch other files
	if tpl == nil {
		tpl = Templates["display/file.html"]
	}

	err := renderTemplate(tpl, pongo2.Context{
		"mime":        metadata.Mimetype,
		"filename":    fileName,
		"size":        sizeHuman,
		"expiry":      expiryHuman,
		"expirylist":  listExpirationTimes(),
		"extra":       extra,
		"forcerandom": Config.forceRandomFilename,
		"lines":       lines,
		"files":       metadata.ArchiveFiles,
		"siteurl":     strings.TrimSuffix(getSiteURL(r), "/"),
	}, r, w)

	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}
