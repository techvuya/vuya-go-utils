package arrayutils

func ArrayStringContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func ArrayChunk[T any](items []T, chunkSize int) [][]T {
	if chunkSize <= 0 {
		return [][]T{} // Return empty slice if chunkSize is invalid
	}

	var chunks [][]T
	for i := 0; i < len(items); i += chunkSize {
		end := i + chunkSize
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}
