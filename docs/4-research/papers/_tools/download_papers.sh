#!/bin/bash
# Ferroelectric CIM Paper Download Script
# Downloads scientific papers from various open-access sources

set -euo pipefail

PAPERS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOWNLOADED_DIR="$PAPERS_DIR/downloaded"
LOGS_DIR="$PAPERS_DIR/logs"
FAILED_LOG="$LOGS_DIR/failed_downloads.txt"
SUCCESS_LOG="$LOGS_DIR/successful_downloads.txt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create directories
mkdir -p "$DOWNLOADED_DIR"/{arxiv,nature,acs,rsc,aip,sciencedirect,springer,frontiers,acm,ieee,chemrxiv,other}
mkdir -p "$LOGS_DIR"

# Initialize logs
echo "=== Download Session: $(date) ===" >> "$FAILED_LOG"
echo "=== Download Session: $(date) ===" >> "$SUCCESS_LOG"

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    echo "[$(date +%H:%M:%S)] $1" >> "$SUCCESS_LOG"
}

log_failed() {
    echo -e "${RED}[FAILED]${NC} $1"
    echo "[$(date +%H:%M:%S)] $1" >> "$FAILED_LOG"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Download with retry and user-agent
download_pdf() {
    local url="$1"
    local output="$2"
    local max_retries=3
    local retry=0

    while [ $retry -lt $max_retries ]; do
        if curl -L -s -f \
            -H "User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" \
            -H "Accept: application/pdf,*/*" \
            --connect-timeout 30 \
            --max-time 120 \
            -o "$output" "$url"; then
            # Check if it's actually a PDF
            if file "$output" | grep -q "PDF"; then
                return 0
            else
                rm -f "$output"
            fi
        fi
        retry=$((retry + 1))
        sleep 2
    done
    return 1
}

# Try Unpaywall API for open access versions
try_unpaywall() {
    local doi="$1"
    local output="$2"
    local email="ferroelectric-cim@example.com"

    log_info "Checking Unpaywall for DOI: $doi"

    local response
    response=$(curl -s "https://api.unpaywall.org/v2/$doi?email=$email" 2>/dev/null || echo "{}")

    local oa_url
    oa_url=$(echo "$response" | jq -r '.best_oa_location.url_for_pdf // empty' 2>/dev/null)

    if [ -n "$oa_url" ] && [ "$oa_url" != "null" ]; then
        log_info "Found open access URL: $oa_url"
        if download_pdf "$oa_url" "$output"; then
            return 0
        fi
    fi
    return 1
}

# Extract arXiv ID from URL or string
get_arxiv_id() {
    local input="$1"
    echo "$input" | grep -oP '(\d{4}\.\d{4,5}(v\d+)?)|([a-z-]+/\d{7})' | head -1
}

# Download from arXiv
download_arxiv() {
    local arxiv_id="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/arxiv/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    local url="https://arxiv.org/pdf/${arxiv_id}.pdf"
    log_info "Downloading arXiv: $arxiv_id"

    if download_pdf "$url" "$output"; then
        log_success "arXiv $arxiv_id -> $filename.pdf"
        return 0
    else
        log_failed "arXiv $arxiv_id"
        return 1
    fi
}

# Download from Nature (try various open access paths)
download_nature() {
    local article_id="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/nature/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    # Try direct PDF URL
    local url="https://www.nature.com/articles/${article_id}.pdf"
    log_info "Trying Nature: $article_id"

    if download_pdf "$url" "$output"; then
        log_success "Nature $article_id -> $filename.pdf"
        return 0
    fi

    # Try to extract DOI and use Unpaywall
    local doi="10.1038/$article_id"
    if try_unpaywall "$doi" "$output"; then
        log_success "Nature $article_id (via Unpaywall) -> $filename.pdf"
        return 0
    fi

    log_failed "Nature $article_id (may require subscription)"
    return 1
}

# Download from ACS Publications
download_acs() {
    local doi_suffix="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/acs/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    local doi="10.1021/$doi_suffix"
    log_info "Trying ACS: $doi"

    # ACS is mostly paywalled, try Unpaywall
    if try_unpaywall "$doi" "$output"; then
        log_success "ACS $doi_suffix (via Unpaywall) -> $filename.pdf"
        return 0
    fi

    log_failed "ACS $doi_suffix (likely paywalled)"
    return 1
}

# Download from RSC (Royal Society of Chemistry)
download_rsc() {
    local article_path="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/rsc/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    log_info "Trying RSC: $article_path"

    # RSC sometimes has open access
    local url="https://pubs.rsc.org/en/content/articlepdf/${article_path}"
    if download_pdf "$url" "$output"; then
        log_success "RSC -> $filename.pdf"
        return 0
    fi

    log_failed "RSC $article_path (may require subscription)"
    return 1
}

# Download from AIP (Journal of Applied Physics, etc)
download_aip() {
    local article_path="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/aip/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    log_info "Trying AIP: $article_path"
    log_failed "AIP $article_path (typically requires subscription)"
    return 1
}

# Download from Frontiers (usually open access)
download_frontiers() {
    local article_path="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/frontiers/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    local url="https://www.frontiersin.org/articles/${article_path}/pdf"
    log_info "Downloading Frontiers: $article_path"

    if download_pdf "$url" "$output"; then
        log_success "Frontiers -> $filename.pdf"
        return 0
    else
        log_failed "Frontiers $article_path"
        return 1
    fi
}

# Download from ChemRxiv (open access preprints)
download_chemrxiv() {
    local article_id="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/chemrxiv/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    # ChemRxiv API endpoint for PDF
    local url="https://chemrxiv.org/engage/api-gateway/chemrxiv/assets/orp/resource/item/${article_id}/original/manuscript.pdf"
    log_info "Downloading ChemRxiv: $article_id"

    if download_pdf "$url" "$output"; then
        log_success "ChemRxiv $article_id -> $filename.pdf"
        return 0
    fi

    # Try alternate URL format
    url="https://chemrxiv.org/engage/chemrxiv/article-details/${article_id}"
    log_failed "ChemRxiv $article_id (may need manual download from: $url)"
    return 1
}

# Download from Springer
download_springer() {
    local article_path="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/springer/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    log_info "Trying Springer: $article_path"

    # Extract DOI and try Unpaywall
    local doi
    doi=$(echo "$article_path" | grep -oP '10\.\d{4,}/[^\s]+' || echo "")

    if [ -n "$doi" ]; then
        if try_unpaywall "$doi" "$output"; then
            log_success "Springer (via Unpaywall) -> $filename.pdf"
            return 0
        fi
    fi

    log_failed "Springer $article_path (may require subscription)"
    return 1
}

# Download from ACM Digital Library
download_acm() {
    local doi_path="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/acm/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    log_info "Trying ACM: $doi_path"

    # Try Unpaywall
    if try_unpaywall "$doi_path" "$output"; then
        log_success "ACM (via Unpaywall) -> $filename.pdf"
        return 0
    fi

    log_failed "ACM $doi_path (may require subscription)"
    return 1
}

# Search Semantic Scholar for papers
search_semantic_scholar() {
    local query="$1"
    local limit="${2:-5}"

    log_info "Searching Semantic Scholar: $query"

    local encoded_query
    encoded_query=$(echo "$query" | sed 's/ /+/g')

    curl -s "https://api.semanticscholar.org/graph/v1/paper/search?query=${encoded_query}&limit=${limit}&fields=title,authors,year,openAccessPdf,externalIds" \
        -H "Accept: application/json" 2>/dev/null | jq -r '.data[] | "\(.title) | \(.year) | \(.openAccessPdf.url // "No PDF")"' 2>/dev/null || echo "Search failed"
}

# Download from Semantic Scholar by paper ID
download_semantic_scholar() {
    local paper_id="$1"
    local filename="$2"
    local output="$DOWNLOADED_DIR/other/${filename}.pdf"

    if [ -f "$output" ]; then
        log_info "Already downloaded: $filename"
        return 0
    fi

    log_info "Fetching from Semantic Scholar: $paper_id"

    local response
    response=$(curl -s "https://api.semanticscholar.org/graph/v1/paper/${paper_id}?fields=openAccessPdf" 2>/dev/null)

    local pdf_url
    pdf_url=$(echo "$response" | jq -r '.openAccessPdf.url // empty' 2>/dev/null)

    if [ -n "$pdf_url" ] && [ "$pdf_url" != "null" ]; then
        if download_pdf "$pdf_url" "$output"; then
            log_success "Semantic Scholar $paper_id -> $filename.pdf"
            return 0
        fi
    fi

    log_failed "Semantic Scholar $paper_id (no open access PDF)"
    return 1
}

# ============================================================================
# PAPER DEFINITIONS
# ============================================================================

download_all_papers() {
    echo ""
    echo "========================================"
    echo "  Ferroelectric CIM Paper Downloader"
    echo "========================================"
    echo ""

    # --- Priority 1: Core HfO2/ZrO2 Ferroelectric Papers ---
    log_info "=== Priority 1: Core HfO2/ZrO2 Papers ==="

    # arXiv papers
    download_arxiv "2401.05288" "first_principles_HfO2_ferroelectric_2024"
    download_arxiv "2307.09357" "ferroelectric_CIM_review_2023"
    download_arxiv "2210.15668" "ferrox_gpu_phasefield_2022"

    # Nature papers
    download_nature "s41467-023-42110-y" "multilevel_fefet_crossbar_2023"
    download_nature "s41598-024-59298-8" "fecap_fefet_cim_elements_2024"
    download_nature "s41467-025-61758-2" "adaptive_control_epitaxial_hfo2_zro2_2025"
    download_nature "s44335-025-00030-8" "dual_bit_fefet_enhanced_storage_2025"

    # --- Priority 2: Compute-in-Memory & Neuromorphic ---
    log_info "=== Priority 2: CIM & Neuromorphic Papers ==="

    download_arxiv "2403.13577" "hcim_adcless_hybrid_cim_2024"
    download_arxiv "2501.06780" "compass_compiler_framework_2025"
    download_arxiv "2411.04814" "simple_packing_algorithm_nvm_2024"

    # --- Priority 3: Physics Models ---
    log_info "=== Priority 3: Physics Models ==="

    # RSC papers
    download_rsc "2024/mh/d4mh00519h" "self_rectifying_ftj_synapse_superlattice_2024"
    download_rsc "2024/nr/d4nr02393e" "2t0c_fedram_multibit_retention_2024"

    # --- Priority 4: Crossbar Non-Idealities ---
    log_info "=== Priority 4: Crossbar Non-Idealities ==="

    download_frontiers "10.3389/femat.2022.988785" "sneak_path_self_rectifying_arrays_2022"
    download_springer "10.1007/s11432-024-4240-x" "hw_sw_codesign_nonidealities_2024"

    # --- Priority 5: ACS Publications ---
    log_info "=== Priority 5: ACS Publications ==="

    download_acs "acsomega.4c10603" "ferroelectric_hfo2_zro2_reduced_wakeup_2024"
    download_acs "acsanm.4c04974" "hfo2_zro2_2d_mos2_nc_transistors_2024"
    download_acs "acsami.3c15732" "ferroelectric_capacitors_superlattice_fatigue_2024"

    # --- Priority 6: ACM Digital Library ---
    log_info "=== Priority 6: ACM Papers ==="

    download_acm "10.1145/3528104" "extreme_partial_sum_quantization_2022"
    download_acm "10.1145/3498328" "dynamic_quantization_range_control_2022"
    download_acm "10.1145/3585518" "variation_tolerant_mapping_2023"

    # --- Priority 7: ChemRxiv ---
    log_info "=== Priority 7: ChemRxiv Preprints ==="

    download_chemrxiv "659ef4cee9ebbb4db9de84cb" "flash_in2se3_neuromorphic_tour_2024"

    # --- Priority 8: Wiley Advanced Science ---
    log_info "=== Priority 8: Wiley Publications ==="

    # These typically need subscription, log for manual download
    log_warn "Wiley papers typically require subscription:"
    log_warn "  - 2D SnS2 Analog Synaptic FeFET: https://advanced.onlinelibrary.wiley.com/doi/10.1002/advs.202308588"
    log_warn "  - Capacitive Crossbar Arrays: https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202100258"

    # --- Priority 9: AIP Publications ---
    log_info "=== Priority 9: AIP Publications ==="

    download_aip "jap/article/136/14/144101/3315849" "oxygen_vacancy_dynamics_superlattice_2024"

    echo ""
    echo "========================================"
    echo "  Download Summary"
    echo "========================================"
    echo ""
    echo "Downloaded PDFs are in: $DOWNLOADED_DIR"
    echo "Success log: $SUCCESS_LOG"
    echo "Failed log: $FAILED_LOG"
    echo ""
    echo "For paywalled papers, consider:"
    echo "  1. Institutional access via university library"
    echo "  2. Contacting authors directly"
    echo "  3. Checking ResearchGate for author uploads"
    echo "  4. Using interlibrary loan services"
    echo ""
}

# ============================================================================
# ADDITIONAL SEARCHES
# ============================================================================

search_tour_papers() {
    echo ""
    log_info "=== Searching for external research group Ferroelectric Papers ==="
    echo ""

    search_semantic_scholar "external research group ferroelectric HfO2" 10
    echo ""
    search_semantic_scholar "external research group neuromorphic computing" 10
    echo ""
    search_semantic_scholar "Jaeho Shin superlattice FeFET" 10
}

search_cim_papers() {
    echo ""
    log_info "=== Searching for Compute-in-Memory Papers ==="
    echo ""

    search_semantic_scholar "ferroelectric compute-in-memory crossbar" 10
    echo ""
    search_semantic_scholar "FeFET analog neural network" 10
    echo ""
    search_semantic_scholar "HfO2 ZrO2 superlattice synapse" 10
}

# ============================================================================
# MAIN
# ============================================================================

print_usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  download    Download all papers (default)"
    echo "  search      Search for additional papers"
    echo "  tour        Search for external research group papers"
    echo "  cim         Search for CIM papers"
    echo "  help        Show this help"
    echo ""
    echo "Examples:"
    echo "  $0                    # Download all papers"
    echo "  $0 download           # Download all papers"
    echo "  $0 search \"query\"     # Search Semantic Scholar"
    echo "  $0 tour               # Search for Tour lab papers"
    echo ""
}

# Check dependencies
check_dependencies() {
    local missing=()

    for cmd in curl jq file; do
        if ! command -v "$cmd" &> /dev/null; then
            missing+=("$cmd")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        echo -e "${RED}Missing dependencies: ${missing[*]}${NC}"
        echo "Install with: sudo apt install ${missing[*]}"
        exit 1
    fi
}

main() {
    check_dependencies

    local command="${1:-download}"

    case "$command" in
        download)
            download_all_papers
            ;;
        search)
            if [ -n "${2:-}" ]; then
                search_semantic_scholar "$2" "${3:-10}"
            else
                echo "Usage: $0 search \"query\" [limit]"
            fi
            ;;
        tour)
            search_tour_papers
            ;;
        cim)
            search_cim_papers
            ;;
        help|--help|-h)
            print_usage
            ;;
        *)
            echo "Unknown command: $command"
            print_usage
            exit 1
            ;;
    esac
}

main "$@"
