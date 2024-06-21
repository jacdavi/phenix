// Code generated by go generate; DO NOT EDIT.
// This file was generated at build time 2024-06-17 23:26:07.504830298 +0000 UTC m=+1.083458634
// This contains all known role checks used in codebase

package rbac

type Permission struct {
	Resource string
	Verb     string
}

var Permissions = []Permission{
	{"applications", "list"},
	{"configs", "create"},
	{"configs", "delete"},
	{"configs", "get"},
	{"configs", "list"},
	{"configs", "update"},
	{"disks", "list"},
	{"exp/captureSubnet", "create"},
	{"experiments", "create"},
	{"experiments", "delete"},
	{"experiments", "get"},
	{"experiments", "list"},
	{"experiments", "patch"},
	{"experiments", "update"},
	{"experiments/apps", "get"},
	{"experiments/captures", "list"},
	{"experiments/files", "get"},
	{"experiments/files", "list"},
	{"experiments/netflow", "create"},
	{"experiments/netflow", "delete"},
	{"experiments/netflow", "get"},
	{"experiments/schedule", "create"},
	{"experiments/schedule", "get"},
	{"experiments/start", "update"},
	{"experiments/stop", "update"},
	{"experiments/topology", "get"},
	{"experiments/trigger", "create"},
	{"experiments/trigger", "delete"},
	{"history", "get"},
	{"hosts", "list"},
	{"miniconsole", "get"},
	{"miniconsole", "post"},
	{"options", "list"},
	{"roles", "list"},
	{"scenarios", "list"},
	{"schemas", "get"},
	{"topologies", "list"},
	{"users", "create"},
	{"users", "delete"},
	{"users", "get"},
	{"users", "list"},
	{"users", "patch"},
	{"users/roles", "patch"},
	{"vms", "delete"},
	{"vms", "get"},
	{"vms", "list"},
	{"vms", "patch"},
	{"vms/captures", "create"},
	{"vms/captures", "delete"},
	{"vms/captures", "list"},
	{"vms/cdrom", "delete"},
	{"vms/cdrom", "update"},
	{"vms/commit", "create"},
	{"vms/forwards", "create"},
	{"vms/forwards", "delete"},
	{"vms/forwards", "get"},
	{"vms/forwards", "list"},
	{"vms/memorySnapshot", "create"},
	{"vms/mount", "delete"},
	{"vms/mount", "get"},
	{"vms/mount", "list"},
	{"vms/mount", "patch"},
	{"vms/mount", "post"},
	{"vms/redeploy", "update"},
	{"vms/reset", "update"},
	{"vms/restart", "update"},
	{"vms/screenshot", "get"},
	{"vms/shutdown", "update"},
	{"vms/snapshots", "create"},
	{"vms/snapshots", "list"},
	{"vms/snapshots", "update"},
	{"vms/start", "update"},
	{"vms/stop", "update"},
	{"vms/vnc", "get"},
	{"workflow", "create"},
}
