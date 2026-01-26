package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"watcher/config"

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

func PDFCompressionHandler(w http.ResponseWriter, r *http.Request) {

	var req PDFCompressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(config.AppConfig.Ghostscript.Path); os.IsNotExist(err) {
		http.Error(w, "Ghostscript binary not found at the configured path", http.StatusInternalServerError)
		return
	}

	if !isValidCompressionLevel(req.CompressionLevel, config.AppConfig.Ghostscript.CompressionLevels) {
		resp := PDFCompressionResponse{
			Message:                    "Invalid compression level",
			AvailableCompressionLevels: config.AppConfig.Ghostscript.CompressionLevels,
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
		AvailableCompressionLevels: config.AppConfig.Ghostscript.CompressionLevels,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func compressPDF(inputPath, outputPath, compressionLevel string) error {
	gsPath := config.AppConfig.Ghostscript.Path

	cmd := exec.Command(gsPath,
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=/"+compressionLevel,
		"-dNOPAUSE",
		"-dQUIET",
		"-dBATCH",
		"-sOutputFile="+outputPath,
		inputPath,
	)

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
