package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

type PDFCompressionRequest struct {
	CompressionLevel string `json:"compressionLevel"`
}

type PDFCompressionResponse struct {
	OutputPath                 string   `json:"outputPath"`
	Message                    string   `json:"message"`
	AvailableCompressionLevels []string `json:"availableCompressionLevels"`
}

var availableLevels = []string{"25", "50", "75", "90", "ghost"}

func PDFCompressionHandler(w http.ResponseWriter, r *http.Request) {
	// new var to store req,body
	var req PDFCompressionRequest

	// decode r.Body store it in the var address
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {

		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !isValidCompressionLevel(req.CompressionLevel, availableLevels) {
		resp := PDFCompressionResponse{
			Message:                    "Invalid compression level",
			AvailableCompressionLevels: availableLevels,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	inputPath := "tmp/pdfcompression/input/input.pdf"
	outputDir := "tmp/pdfcompression/output"
	outputFileName := uuid.New().String() + ".pdf"
	outputPath := filepath.Join(outputDir, outputFileName)

	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		http.Error(w, "Input file not found", http.StatusNotFound)
		return
	}

	if err := compressPDF(inputPath, outputPath, req.CompressionLevel); err != nil {
		http.Error(w, fmt.Sprintf("Failed to compress PDF: %v", err), http.StatusInternalServerError)
		return
	}

	resp := PDFCompressionResponse{
		OutputPath:                 outputPath,
		Message:                    "PDF compressed successfully",
		AvailableCompressionLevels: availableLevels,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getCompressionArgs(level string) []string {
	levels := map[string][]string{
		"25": {
			"-sDEVICE=pdfwrite",
			"-dCompatibilityLevel=1.4",
			"-dDownsampleColorImages=true",
			"-dDownsampleGrayImages=true",
			"-dDownsampleMonoImages=true",
			"-dColorImageResolution=150",
			"-dGrayImageResolution=150",
			"-dMonoImageResolution=150",
			"-dColorImageDownsampleType=/Bicubic",
			"-dGrayImageDownsampleType=/Bicubic",
			"-dMonoImageDownsampleType=/Subsample",
			"-dJPEGQ=85",
			"-dNOPAUSE",
			"-dQUIET",
			"-dBATCH",
		},
		"50": {
			"-sDEVICE=pdfwrite",
			"-dCompatibilityLevel=1.4",
			"-dDownsampleColorImages=true",
			"-dDownsampleGrayImages=true",
			"-dDownsampleMonoImages=true",
			"-dColorImageResolution=120",
			"-dGrayImageResolution=120",
			"-dMonoImageResolution=120",
			"-dColorImageDownsampleType=/Bicubic",
			"-dGrayImageDownsampleType=/Bicubic",
			"-dMonoImageDownsampleType=/Subsample",
			"-dJPEGQ=70",
			"-dNOPAUSE",
			"-dQUIET",
			"-dBATCH",
		},
		"75": {
			"-sDEVICE=pdfwrite",
			"-dCompatibilityLevel=1.4",
			"-dDownsampleColorImages=true",
			"-dDownsampleGrayImages=true",
			"-dDownsampleMonoImages=true",
			"-dColorImageResolution=100",
			"-dGrayImageResolution=100",
			"-dMonoImageResolution=100",
			"-dColorImageDownsampleType=/Bicubic",
			"-dGrayImageDownsampleType=/Bicubic",
			"-dMonoImageDownsampleType=/Subsample",
			"-dJPEGQ=60",
			"-dNOPAUSE",
			"-dQUIET",
			"-dBATCH",
		},
		"90": {
			"-sDEVICE=pdfwrite",
			"-dCompatibilityLevel=1.4",
			"-dDownsampleColorImages=true",
			"-dDownsampleGrayImages=true",
			"-dDownsampleMonoImages=true",
			"-dColorImageResolution=72",
			"-dGrayImageResolution=72",
			"-dMonoImageResolution=72",
			"-dColorImageDownsampleType=/Bicubic",
			"-dGrayImageDownsampleType=/Bicubic",
			"-dMonoImageDownsampleType=/Subsample",
			"-dJPEGQ=45",
			"-dNOPAUSE",
			"-dQUIET",
			"-dBATCH",
		},
		"ghost": {
			"-dPDFSETTINGS=/ebook",
			"-dDownsampleColorImages=true",
			"-dDownsampleGrayImages=true",
			"-dDownsampleMonoImages=true",
			"-dColorImageResolution=100",
			"-dGrayImageResolution=100",
			"-dMonoImageResolution=100",
			"-dColorImageDownsampleType=/Bicubic",
			"-dGrayImageDownsampleType=/Bicubic",
			"-dMonoImageDownsampleType=/Subsample",
			"-dJPEGQ=60",
			"-dAutoRotatePages=/None",
		},
	}
	return levels[level]
}

func compressPDF(inputPath, outputPath, compressionLevel string) error {
	gsPath := "gsc"
	args := getCompressionArgs(compressionLevel)
	args = append(args, "-sOutputFile="+outputPath, inputPath)

	cmd := exec.Command(gsPath, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %v, output: %s", err, string(output))
	}

	return nil
}

func isValidCompressionLevel(level string, availableLevels []string) bool {
	for _, l := range availableLevels {
		if l == level {
			return true
		}
	}
	return false
}
