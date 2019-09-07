package processor

import (
	"github.com/ddouglas/monocle"
)

func chunkCharacterSlice(size int, slice []monocle.Character) [][]monocle.Character {

	var chunk [][]monocle.Character
	chunk = make([][]monocle.Character, 0)

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
