package cloudkit

import (
	"errors"

	"github.com/digitalocean/go-libvirt"
)

// MemStats is a human readable version of the []libvirt.DomainMemoryStat that comes
// back from libvirt's DomainMemoryStats func.
type MemStats struct {
	Actual     uint64
	SwapIn     uint64
	SwapOut    uint64
	MajorFault uint64
	MinorFault uint64
	Unused     uint64
	Available  uint64
	Usable     uint64
	LastUpdate uint64
	Rss        uint64
}

// NewMemStats is a custom unmarshaller for the array of array of numbers we get
// when asking libvirt for memory statistics on a domain
func NewMemStats(data []libvirt.DomainMemoryStat) (MemStats, error) {
	if len(data) != 10 {
		return MemStats{}, errors.New("debug MemStats")
	}
	return MemStats{
		Actual:     data[0].Val,
		SwapIn:     data[1].Val,
		SwapOut:    data[2].Val,
		MajorFault: data[3].Val,
		MinorFault: data[4].Val,
		Unused:     data[5].Val,
		Available:  data[6].Val,
		Usable:     data[7].Val,
		LastUpdate: data[8].Val,
		Rss:        data[9].Val,
	}, nil
}

// rStats come back as a []libvirt.DomainMemoryStat... They all have a "Tag" field as
// well as their associated Val. I can't for the life of me find anywhere that talks
// about what the tag means. I initially imagined the data could maybe come back in
// any order or something, but I could not prove that once ever (out of hundreds).
// I ended up figuring out which values meant what by comparing them to the output
// of `virsh dommemstat {domain}`... Thanksfully they were in the same order from
// top to bottom and I labeled them below. I'll come back to this. I'm sure I'll
// have to as soon as I add different machines, as I could see this varying.

// rStats: [
// 	{
// 		Tag: 6
// 		Val: 2097152 // actual
// 	},{
// 		Tag: 0
// 		Val: 0 // swap_in
// 	},{
// 		Tag: 1
// 		Val: 0 // swap_out
// 	},{
// 		Tag: 2
// 		Val: 922 // major_fault
// 	},{
// 		Tag: 3
// 		Val: 314341 // minor_fault
// 	},{
// 		Tag: 4
// 		Val: 1787532 // unused
// 	},{
// 		Tag: 5
// 		Val: 2041024 // available
// 	},{
// 		Tag: 8
// 		Val: 1830488 // usable
// 	},{
// 		Tag: 9
// 		Val: 1605988199 // last_update
// 	},{
// 		Tag: 7
// 		Val: 457240 // rssx`
// 	}
// ]
