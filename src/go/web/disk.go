package web

import (
	"encoding/json"
	"net/http"
	"phenix/api/cluster"
	"phenix/util/plog"
	"phenix/web/rbac"
	"phenix/web/util"
	"sort"
	"strings"
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
		defaultDiskType = cluster.VM_IMAGE | cluster.CONTAINER_IMAGE | cluster.ISO_IMAGE | cluster.UNKNOWN
	)

	if !role.Allowed("disks", "list") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if len(diskType) > 0 {
		defaultDiskType = 0
		for  _, s := range strings.Split(diskType, ",") {
			defaultDiskType |= cluster.StringToKind(s)
		}
	
	}

	disks, err := cluster.GetImages(expName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filtered := []cluster.ImageDetails{}
	for _, disk := range disks {
		if disk.Kind&defaultDiskType != 0 {
			filtered = append(filtered, disk)
		}
	}
	
	allowed := []cluster.ImageDetails{}
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
