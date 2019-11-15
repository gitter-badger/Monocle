package tools

import (
	"fmt"

	"github.com/ddouglas/monocle"
	"github.com/volatiletech/sqlboiler/queries"
)

func OutputDebugQuery(query *queries.Query) {

	stmt, args := queries.BuildQuery(query)
	fmt.Printf("\n\nQuery: %s\nArgs: %v\n\n", stmt, args)

}

func CharIdsFromSlice(characters []monocle.Character) []uint64 {

	ids := []uint64{}

	for _, c := range characters {
		ids = append(ids, c.ID)
	}

	return ids
}

func CharIDsFromCharMap(characters map[uint64]monocle.Character) []uint64 {

	ids := []uint64{}
	for id := range characters {
		ids = append(ids, id)
	}

	return ids

}

func CharacterSliceToMap(characters []monocle.Character) map[uint64]monocle.Character {

	chunk := make(map[uint64]monocle.Character)
	for _, c := range characters {
		chunk[c.ID] = c
	}

	return chunk
}

func ChunkCharacterSlice(size int, slice []monocle.Character) [][]monocle.Character {

	var chunk [][]monocle.Character
	chunk = make([][]monocle.Character, 0)

	if len(slice) <= size {
		chunk = append(chunk, slice)
		return chunk
	}

	for x := 0; x <= len(slice)-1; x += size {

		end := x + size

		if end > len(slice) {
			end = len(slice)
		}

		chunk = append(chunk, slice[x:end])

	}

	return chunk
}

func ChunkCorporationSlice(size int, slice []monocle.Corporation) [][]monocle.Corporation {

	var chunk [][]monocle.Corporation
	chunk = make([][]monocle.Corporation, 0)

	if len(slice) <= size {
		chunk = append(chunk, slice)
		return chunk
	}

	for x := 0; x <= len(slice)-1; x += size {

		end := x + size

		if end > len(slice) {
			end = len(slice)
		}

		chunk = append(chunk, slice[x:end])

	}

	return chunk
}

func CorpIDsFromCorpMap(corporations map[uint]monocle.Corporation) []uint {

	ids := []uint{}
	for id := range corporations {
		ids = append(ids, id)
	}

	return ids

}

func CorporationSliceToMap(corporations []monocle.Corporation) map[uint]monocle.Corporation {

	chunk := make(map[uint]monocle.Corporation)
	for _, c := range corporations {
		chunk[c.ID] = c
	}

	return chunk
}

func ChunkAllianceSlice(size int, slice []monocle.Alliance) [][]monocle.Alliance {

	var chunk [][]monocle.Alliance
	chunk = make([][]monocle.Alliance, 0)

	if len(slice) <= size {
		chunk = append(chunk, slice)
		return chunk
	}

	for x := 0; x <= len(slice); x += size {

		end := x + size

		if end > len(slice) {
			end = len(slice)
		}

		chunk = append(chunk, slice[x:end])

	}

	return chunk
}
