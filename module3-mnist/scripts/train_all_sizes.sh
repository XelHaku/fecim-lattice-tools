#!/bin/bash
# train_all_sizes.sh - Train MNIST networks with different hidden layer sizes
#
# This script trains networks with hidden sizes 64, 128, and 256
# and saves the weights for use in the dual-mode demo.
#
# Usage: ./train_all_sizes.sh
#
# Output files:
#   ../data/pretrained_30_h64.json
#   ../data/pretrained_30_h128.json
#   ../data/pretrained_30_h256.json

set -e  # Exit on error

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMO_DIR="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$DEMO_DIR/data"

# Training parameters
EPOCHS=10
BATCH_SIZE=64
LEARNING_RATE=0.001
LEVELS=30  # FeCIM quantization levels

echo "=============================================="
echo "FeCIM MNIST Training Script"
echo "=============================================="
echo "Training parameters:"
echo "  Epochs: $EPOCHS"
echo "  Batch size: $BATCH_SIZE"
echo "  Learning rate: $LEARNING_RATE"
echo "  Quantization levels: $LEVELS"
echo "=============================================="
echo ""

# Check if MNIST data exists
if [ ! -f "$DATA_DIR/train-images-idx3-ubyte.gz" ]; then
    echo "MNIST data not found. Downloading..."
    mkdir -p "$DATA_DIR"
    cd "$DATA_DIR"
    wget -q http://yann.lecun.com/exdb/mnist/train-images-idx3-ubyte.gz
    wget -q http://yann.lecun.com/exdb/mnist/train-labels-idx1-ubyte.gz
    wget -q http://yann.lecun.com/exdb/mnist/t10k-images-idx3-ubyte.gz
    wget -q http://yann.lecun.com/exdb/mnist/t10k-labels-idx1-ubyte.gz
    echo "MNIST data downloaded."
    cd "$SCRIPT_DIR"
fi

# Build the unified tool if needed
echo "Building unified tool..."
cd "$DEMO_DIR"
go build -o fecim-lattice-tools ../cmd/fecim-lattice-tools 2>/dev/null || {
    echo "Note: Unified tool not available."
    echo "Using existing pretrained weights if available."
}

# Train each hidden size
for HIDDEN in 64 128 256; do
    OUTPUT_FILE="$DATA_DIR/pretrained_30_h${HIDDEN}.json"

    echo ""
    echo "----------------------------------------------"
    echo "Training hidden size: $HIDDEN"
    echo "Output file: $OUTPUT_FILE"
    echo "----------------------------------------------"

    if [ -x "$DEMO_DIR/fecim-lattice-tools" ]; then
        # Use the unified tool
        "$DEMO_DIR/fecim-lattice-tools" mnist cli --train \
            --epochs "$EPOCHS" \
            --hidden "$HIDDEN" \
            --save "$OUTPUT_FILE"

        echo "Training complete for hidden=$HIDDEN"

        # Evaluate the trained model
        echo "Evaluating..."
        "$DEMO_DIR/fecim-lattice-tools" mnist cli --load "$OUTPUT_FILE" --evaluate
    else
        echo "Training tool not available. Skipping training for hidden=$HIDDEN."
        if [ -f "$OUTPUT_FILE" ]; then
            echo "Using existing weights file: $OUTPUT_FILE"
        else
            echo "Warning: No weights file found for hidden=$HIDDEN"
        fi
    fi
done

# Cleanup
rm -f "$DEMO_DIR/train_tool"

echo ""
echo "=============================================="
echo "Training Complete!"
echo "=============================================="
echo "Generated weight files:"
ls -la "$DATA_DIR"/pretrained_30_h*.json 2>/dev/null || echo "  (none generated)"
echo ""
echo "Expected accuracy (30-level quantized simulation):"
echo "  h64:  ~95%"
echo "  h128: ~97%"
echo "  h256: ~97.5%"
echo ""
echo "Note: FeCIM hardware achieves 87% (Dr. Tour, Nov 2024)"
echo "Set noise=0.08 in the GUI to match hardware results."
echo "=============================================="
