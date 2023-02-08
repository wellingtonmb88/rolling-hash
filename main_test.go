package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSignature(t *testing.T) {
	type expected struct {
		fingerprint []byte
		err         error
	}
	testCases := []struct {
		name     string
		input    string
		expected expected
	}{
		{
			name:  "should validate create fingerprint successfully",
			input: "test_files/file_original.txt",
			expected: expected{
				fingerprint: []byte{47, 34, 62, 47, 12, 34, 78, 34, 26},
				err:         nil,
			},
		},
		{
			name:  "should fail with wrong original file path",
			input: "file_missing.txt",
			expected: expected{
				err: errors.New("open file_missing.txt: no such file or directory"),
			},
		},
	}

	for _, test := range testCases {
		err := Signature(test.input, "fingerprint.txt")

		if test.expected.err != nil {
			require.NotNil(t, err)
			require.Equal(t, test.expected.err.Error(), err.Error(), test.name)
			continue
		}

		require.Nil(t, err)
		fingerprint, err := readFromFile("fingerprint.txt")
		require.Nil(t, err)
		require.NotNil(t, fingerprint)
		require.Equal(t, test.expected.fingerprint, fingerprint, test.name)
	}
}

func TestDelta(t *testing.T) {
	err := Signature("test_files/file_original.txt", "fingerprint.txt")
	require.Nil(t, err)
	fingerprint := []byte{47, 34, 62, 47, 12, 34, 78, 34, 26}

	type expected struct {
		DeltaData
		err error
	}
	type input struct {
		fingerprintPath string
		filePath        string
	}
	testCases := []struct {
		name     string
		input    input
		expected expected
	}{
		{
			name:  "should validate fingerprint with file adding at the begining successfully",
			input: input{"fingerprint.txt", "test_files/file_add_begining.txt"},
			expected: expected{
				DeltaData: DeltaData{
					Fingerprint:   fingerprint,
					Chunks:        []byte{94, 63, 47, 34, 62, 47, 12, 34, 78, 34, 26},
					NewChunks:     []byte{94, 63},
					ChangedChunks: []byte{},
					DeletedChunks: []byte{},
				},
				err: nil,
			},
		},
		{
			name:  "should validate fingerprint with file adding at the end successfully",
			input: input{"fingerprint.txt", "test_files/file_add_end.txt"},
			expected: expected{
				DeltaData: DeltaData{
					Fingerprint:   fingerprint,
					Chunks:        []byte{47, 34, 62, 47, 12, 34, 78, 34, 26, 74},
					NewChunks:     []byte{74},
					ChangedChunks: []byte{},
					DeletedChunks: []byte{},
				},
				err: nil,
			},
		},
		{
			name:  "should validate fingerprint with file adding in the middle successfully",
			input: input{"fingerprint.txt", "test_files/file_add_middle.txt"},
			expected: expected{
				DeltaData: DeltaData{
					Fingerprint:   fingerprint,
					Chunks:        []byte{47, 34, 62, 47, 12, 79, 59, 81, 34, 26},
					NewChunks:     []byte{79, 59, 81},
					ChangedChunks: []byte{},
					DeletedChunks: []byte{},
				},
				err: nil,
			},
		},
		{
			name:  "should validate fingerprint with file changing at the end successfully",
			input: input{"fingerprint.txt", "test_files/file_change_end.txt"},
			expected: expected{
				DeltaData: DeltaData{
					Fingerprint:   fingerprint,
					Chunks:        []byte{47, 34, 62, 47, 12, 34, 78, 34, 80},
					NewChunks:     []byte{},
					ChangedChunks: []byte{80},
					DeletedChunks: []byte{},
				},
				err: nil,
			},
		},
		{
			name:  "should validate fingerprint with file deleting content successfully",
			input: input{"fingerprint.txt", "test_files/file_delete.txt"},
			expected: expected{
				DeltaData: DeltaData{
					Fingerprint:   fingerprint,
					Chunks:        []byte{47, 34, 62, 47, 12, 34, 78},
					NewChunks:     []byte{},
					ChangedChunks: []byte{},
					DeletedChunks: []byte{34, 26},
				},
				err: nil,
			},
		},
		{
			name:  "should validate fingerprint with original file content successfully",
			input: input{"fingerprint.txt", "test_files/file_original.txt"},
			expected: expected{
				DeltaData: DeltaData{
					Fingerprint:   fingerprint,
					Chunks:        fingerprint,
					NewChunks:     []byte{},
					ChangedChunks: []byte{},
					DeletedChunks: []byte{},
				},
				err: nil,
			},
		},
		{
			name:  "should fail with wrong file path",
			input: input{"fingerprint.txt", "file_missing.txt"},
			expected: expected{
				err: errors.New("open file_missing.txt: no such file or directory"),
			},
		},
		{
			name:  "should fail with wrong fingerprint file path",
			input: input{"fingerprint_missing.txt", "test_files/file_original.txt"},
			expected: expected{
				err: errors.New("open fingerprint_missing.txt: no such file or directory"),
			},
		},
	}

	for _, test := range testCases {
		data, err := Delta(test.input.fingerprintPath, test.input.filePath)

		if test.expected.err != nil {
			require.NotNil(t, err)
			require.Equal(t, test.expected.err.Error(), err.Error(), test.name)
			continue
		}

		require.Nil(t, err)
		require.NotNil(t, data)

		require.Equal(t, test.expected.Fingerprint, data.Fingerprint, test.name)
		require.Equal(t, test.expected.Chunks, data.Chunks, test.name)
		require.Equal(t, test.expected.NewChunks, data.NewChunks, test.name)
		require.Equal(t, test.expected.ChangedChunks, data.ChangedChunks, test.name)
		require.Equal(t, test.expected.DeletedChunks, data.DeletedChunks, test.name)
	}
}
