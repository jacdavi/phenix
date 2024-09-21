package web

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"phenix/api/disk"
	"phenix/util/plog"
	"phenix/web/rbac"
	"phenix/web/util"
	"sort"
	"strings"

	"github.com/gorilla/mux"
)

// GET /disks
func GetDisks(w http.ResponseWriter, r *http.Request) {
	plog.Debug("HTTP handler called", "handler", "GetDisks")

	var (
		ctx             = r.Context()
		role            = ctx.Value("role").(rbac.Role)
		query           = r.URL.Query()
		expName         = query.Get("expName")
		diskType        = query.Get("diskType")
		defaultDiskType = disk.VM_IMAGE | disk.CONTAINER_IMAGE | disk.ISO_IMAGE | disk.UNKNOWN
	)

	if !role.Allowed("disks", "list") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if len(diskType) > 0 {
		defaultDiskType = 0
		for  _, s := range strings.Split(diskType, ",") {
			defaultDiskType |= disk.StringToKind(s)
		}
	
	}

	disks, err := disk.GetImages(expName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filtered := []disk.Details{}
	for _, disk := range disks {
		if disk.Kind&defaultDiskType != 0 {
			filtered = append(filtered, disk)
		}
	}
	
	allowed := []disk.Details{}
	for _, disk := range filtered {
		if role.Allowed("disks", "list", disk.Name) {
			allowed = append(allowed, disk)
		}
	}

	sort.Slice(allowed, func(i, j int) bool {
		return allowed[i].Name < allowed[j].Name
	})

	body, err := json.Marshal(util.WithRoot("disks", allowed))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

// POST /disks/commit?path={path}
func CommitDisk(w http.ResponseWriter, r *http.Request) {
	plog.Debug("HTTP handler called", "handler", "CommitDisk")
	role := r.Context().Value("role").(rbac.Role)
	path := mux.Vars(r)["path"]
	plog.Info("got", "disk", path)

	if !role.Allowed("disks", "post", path[strings.LastIndex(path, "/")+1:]) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	err := disk.CommitDisk(path)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}


// POST /disks/snapshot?src={src}&dst={dst}
// src should be absolute
// dst may be absolute, but will be put in same dir if not. Extension will match src if not provided
func SnapshotDisk(w http.ResponseWriter, r *http.Request) {
	plog.Debug("HTTP handler called", "handler", "CommitDisk")
	role := r.Context().Value("role").(rbac.Role)
	src := mux.Vars(r)["src"]
	dst := mux.Vars(r)["dst"]

	if !filepath.IsAbs(dst) {
		dst = filepath.Join(filepath.Dir(src), dst)
	}

	if filepath.Ext(dst) == "" {
		dst = dst + filepath.Ext(src)
	}

	if !role.Allowed("disks", "post", dst[strings.LastIndex(dst, "/")+1:]) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	err := disk.SnapshotDisk(src, dst)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}