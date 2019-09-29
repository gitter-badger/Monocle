package processor

import (
	"github.com/ddouglas/monocle"
)

func charIdsFromSlice(characters []monocle.Character) []uint64 {

	ids := []uint64{}

	for _, c := range characters {
		ids = append(ids, c.ID)
	}

	return ids
}

func characterSliceToMap(characters []monocle.Character) map[uint64]monocle.Character {

	chunk := make(map[uint64]monocle.Character, 0)
	for _, c := range characters {
		chunk[c.ID] = c
	}

	return chunk

}

func chunkCharacterSlice(size int, slice []monocle.Character) [][]monocle.Character {

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

func chunkCorporationSlice(size int, slice []monocle.Corporation) [][]monocle.Corporation {

	var chunk [][]monocle.Corporation
	chunk = make([][]monocle.Corporation, 0)

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
