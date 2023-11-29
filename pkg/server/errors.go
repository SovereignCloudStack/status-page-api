package server

import "errors"

var (
	// ErrInvalidPhaseGeneration means the given generation is invalid.
	// This can be seen as 400 - Bad request.
	ErrInvalidPhaseGeneration = errors.New("invalid phase generation")

	// ErrPhaseGenerationNotFound means the given generation was not found.
	// this can be seen as 404 - Not found.
	ErrPhaseGenerationNotFound = errors.New("phase generation not found")
)
