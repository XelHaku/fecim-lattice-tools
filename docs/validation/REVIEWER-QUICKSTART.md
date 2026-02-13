# Reviewer Quick Start (5 minutes)

## Prerequisites
- Go 1.24+ installed
- Clone: `git clone https://github.com/your-org/fecim-lattice-tools.git`

## Verify
cd fecim-lattice-tools
bash scripts/reproduce_validation.sh

## Expected Output
- BUILD: PASS
- VET: PASS
- 70 packages pass / 0 fail
- Physics regression: within tolerance
- Kirchhoff: residuals < 1e-12
- ISPP: 0 spurious discontinuities
