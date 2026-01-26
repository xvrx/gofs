// compress.js
const { exec } = require("child_process");








let gsBinary = "gsc"; // placeholder (user can override)

export function setGsBinary(path: string) {
    /**
  * Set the Ghostscript binary path.
  * Example:
  *   setGsBinary("/usr/local/bin/gs")
  */
    gsBinary = path;
}

/**
 * Verify if Ghostscript is installed and accessible.
 * Returns true if valid, otherwise throws.
 */
export async function verifyGhostscript() {
    try {
        await exec(`"${gsBinary}" -v`);
        console.log("GhostSript binary confirmed")
        return true;
    } catch (err) {
        throw new Error(
            `Ghostscript not found at: ${gsBinary}\n` +
            `Please install Ghostscript or set binary path via setGsBinary("/path/to/gs").`
        );
    }
}

/**
 * Run Ghostscript with given args
 */
export function runGhostscript(args: string[]) {
    const cmd = `${gsBinary} ${args.join(" ")}`

    console.log("❗executing cmd :")
    console.log(cmd)

    exec(cmd, (err: any, stdout: any, stderr: any) => {
        if (err) {
            console.log("error while executing running script")
            console.log(err)
            return;
        } else {
            console.log("✅ No error found")
            console.log(`stdout: ${stdout}`);
            console.log(`stderr: ${stderr}`);
        }
    })
}



type compressionParams = {
    inputDirectory: string,
    inputFileName: string,
    outputDirectory: string,
    outputFileName: string,
    level?: "25" | "50" | "75" | "90" | "ghost"
}

/**
 * Compress a PDF
 * @param {string} input - Path to input PDF
 * @param {string} output - Path to output PDF
 * @param {"25"|"50"|"75"|"90"} level
 */


export async function compressPDF({
    inputDirectory = "./",
    outputDirectory = "./",
    inputFileName,
    outputFileName,
    level = "25"
}: compressionParams) {

    await verifyGhostscript();

    const baseArgs = [
        "-sDEVICE=pdfwrite",
        "-dCompatibilityLevel=1.4",
        "-dNOPAUSE",
        "-dQUIET",
        "-dBATCH",
        "-sProducer=Unknown",
        "-sCreator=Unknown"
    ];

    interface type_compression_level {
        [key: string]: string[]; // Defines that any string key will map to a string value
    }

    const levels: type_compression_level = {
        "25": [
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
            "-dBATCH"
        ],
        "50": [
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
            "-dBATCH"
        ],
        "75": [
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
            "-dBATCH"
        ],
        "90": [
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
            "-dBATCH"
        ],
        "ghost": [
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
            "-dAutoRotatePages=/None"
        ]
    };

    const args = [
        baseArgs.join(" "),
        (levels[level]).join(" "),
        // replace node js backslash with bash readable slash
        `-sOutputFile=${outputDirectory.replace(/\\/g, "/")}/${outputFileName}.pdf`,
        `${inputDirectory.replace(/\\/g, "/")}/${inputFileName}`,
    ];

    return runGhostscript(args);
}
