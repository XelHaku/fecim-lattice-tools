#!/usr/bin/env python3
"""Brevitas QAT vs Go PTQ comparison for MNIST 5-bit/30-level quantization.

Trains a simple MNIST MLP (784->128->10) to full precision, then compares:
  - Full-precision (FP32) accuracy
  - Post-training quantization (PTQ) at 30 levels (simulates Go quantize.go)
  - Brevitas quantization-aware training (QAT) at 5-bit weights (~32 levels)

Outputs JSON on stdout with accuracies and weight statistics.

Usage:
    pip install torch torchvision brevitas
    python3 scripts/brevitas_qat_compare.py

The script is self-contained and downloads MNIST data to /tmp/mnist_data.
"""

import json
import sys
import math

import torch
import torch.nn as nn
import torch.optim as optim
from torchvision import datasets, transforms

# Brevitas imports for QAT
try:
    import brevitas.nn as qnn
    from brevitas.quant import Int8WeightPerTensorFloat
    BREVITAS_AVAILABLE = True
except ImportError:
    BREVITAS_AVAILABLE = False
    print(json.dumps({"error": "brevitas not installed. pip install brevitas"}))
    sys.exit(1)


# --- Model definitions ---

class FPModel(nn.Module):
    """Full-precision MNIST MLP: 784 -> 128 -> 10."""
    def __init__(self):
        super().__init__()
        self.fc1 = nn.Linear(784, 128)
        self.relu = nn.ReLU()
        self.fc2 = nn.Linear(128, 10)

    def forward(self, x):
        x = x.view(-1, 784)
        x = self.relu(self.fc1(x))
        x = self.fc2(x)
        return x


class QATModel(nn.Module):
    """Brevitas QAT MNIST MLP with 5-bit weights (~30 levels).

    Uses QuantLinear with 5-bit weight quantization to match the FeCIM
    30-level baseline (2^5 = 32 levels, close to 30 = 2^4.91).
    """
    def __init__(self):
        super().__init__()
        self.fc1 = qnn.QuantLinear(
            784, 128,
            bias=True,
            weight_bit_width=5,
        )
        self.relu = qnn.QuantReLU(bit_width=8)
        self.fc2 = qnn.QuantLinear(
            128, 10,
            bias=True,
            weight_bit_width=5,
        )

    def forward(self, x):
        x = x.view(-1, 784)
        x = self.relu(self.fc1(x))
        x = self.fc2(x)
        return x


# --- PTQ simulation (matches Go QuantizeWeights logic) ---

def ptq_quantize_tensor(tensor, levels=30):
    """Quantize a weight tensor to N symmetric levels, matching Go implementation.

    Maps [-wmax, +wmax] -> integer bins [0, levels-1] -> back to float.
    This is the same algorithm as shared/neural/quantize.go:QuantizeWeights.
    """
    w = tensor.detach().clone()
    w_max = w.abs().max().item()
    if w_max == 0:
        return w

    level_step = 2.0 * w_max / (levels - 1)
    norm_scale = (levels - 1) / (2.0 * w_max)

    # Map to bins and back
    bins = torch.round((w + w_max) * norm_scale).clamp(0, levels - 1)
    quantized = -w_max + bins * level_step
    return quantized


def apply_ptq(model, levels=30):
    """Apply PTQ to a copy of the model (same algo as Go quantize.go)."""
    model_copy = FPModel()
    model_copy.load_state_dict(model.state_dict())
    with torch.no_grad():
        model_copy.fc1.weight.copy_(ptq_quantize_tensor(model_copy.fc1.weight, levels))
        model_copy.fc1.bias.copy_(ptq_quantize_tensor(model_copy.fc1.bias, levels))
        model_copy.fc2.weight.copy_(ptq_quantize_tensor(model_copy.fc2.weight, levels))
        model_copy.fc2.bias.copy_(ptq_quantize_tensor(model_copy.fc2.bias, levels))
    return model_copy


# --- Training & evaluation ---

def train_model(model, train_loader, epochs=5, lr=0.001):
    """Train a model on MNIST."""
    criterion = nn.CrossEntropyLoss()
    optimizer = optim.Adam(model.parameters(), lr=lr)

    model.train()
    for epoch in range(epochs):
        running_loss = 0.0
        for batch_idx, (data, target) in enumerate(train_loader):
            optimizer.zero_grad()
            output = model(data)
            loss = criterion(output, target)
            loss.backward()
            optimizer.step()
            running_loss += loss.item()

        avg_loss = running_loss / len(train_loader)
        print(f"  Epoch {epoch+1}/{epochs}, loss: {avg_loss:.4f}", file=sys.stderr)

    return model


def evaluate_model(model, test_loader):
    """Evaluate model accuracy on MNIST test set."""
    model.eval()
    correct = 0
    total = 0
    with torch.no_grad():
        for data, target in test_loader:
            output = model(data)
            _, predicted = output.max(1)
            correct += predicted.eq(target).sum().item()
            total += target.size(0)
    return correct / total


def weight_stats(tensor):
    """Compute weight statistics for a tensor."""
    w = tensor.detach().cpu()
    return {
        "mean": w.mean().item(),
        "std": w.std().item(),
        "min": w.min().item(),
        "max": w.max().item(),
        "num_unique": len(w.unique()),
        "shape": list(w.shape),
    }


def main():
    print("Loading MNIST dataset...", file=sys.stderr)

    transform = transforms.Compose([
        transforms.ToTensor(),
        transforms.Normalize((0.1307,), (0.3081,)),
    ])

    train_dataset = datasets.MNIST('/tmp/mnist_data', train=True, download=True, transform=transform)
    test_dataset = datasets.MNIST('/tmp/mnist_data', train=False, download=True, transform=transform)

    train_loader = torch.utils.data.DataLoader(train_dataset, batch_size=64, shuffle=True)
    test_loader = torch.utils.data.DataLoader(test_dataset, batch_size=1000, shuffle=False)

    # --- Phase 1: Train full-precision model ---
    print("\n--- Training full-precision model ---", file=sys.stderr)
    fp_model = FPModel()
    fp_model = train_model(fp_model, train_loader, epochs=5, lr=0.001)
    fp_accuracy = evaluate_model(fp_model, test_loader)
    print(f"FP accuracy: {fp_accuracy*100:.2f}%", file=sys.stderr)

    # --- Phase 2: Apply PTQ at 30 levels (same as Go) ---
    print("\n--- Applying PTQ at 30 levels ---", file=sys.stderr)
    ptq_model = apply_ptq(fp_model, levels=30)
    ptq_accuracy = evaluate_model(ptq_model, test_loader)
    print(f"PTQ accuracy (30 levels): {ptq_accuracy*100:.2f}%", file=sys.stderr)

    # --- Phase 3: Train QAT model with 5-bit weights ---
    print("\n--- Training QAT model (5-bit weights) ---", file=sys.stderr)
    qat_model = QATModel()

    # Initialize QAT model from FP weights for fair comparison
    fp_state = fp_model.state_dict()
    qat_state = qat_model.state_dict()
    for key in ['fc1.weight', 'fc1.bias', 'fc2.weight', 'fc2.bias']:
        if key in qat_state and key in fp_state:
            qat_state[key] = fp_state[key]
    qat_model.load_state_dict(qat_state, strict=False)

    # Fine-tune with QAT (lower LR since we start from pre-trained)
    qat_model = train_model(qat_model, train_loader, epochs=5, lr=0.0005)
    qat_accuracy = evaluate_model(qat_model, test_loader)
    print(f"QAT accuracy (5-bit): {qat_accuracy*100:.2f}%", file=sys.stderr)

    # --- Collect weight statistics ---
    fp_w1_stats = weight_stats(fp_model.fc1.weight)
    ptq_w1_stats = weight_stats(ptq_model.fc1.weight)

    # For QAT, get the quantized weight values
    qat_w1_stats = weight_stats(qat_model.fc1.weight)

    # --- Output JSON result ---
    result = {
        "fp_accuracy": round(fp_accuracy, 6),
        "ptq_accuracy": round(ptq_accuracy, 6),
        "qat_accuracy": round(qat_accuracy, 6),
        "ptq_levels": 30,
        "qat_bit_width": 5,
        "qat_levels": 32,  # 2^5
        "accuracy_delta_qat_vs_ptq": round(qat_accuracy - ptq_accuracy, 6),
        "accuracy_delta_ptq_vs_fp": round(ptq_accuracy - fp_accuracy, 6),
        "weight_stats": {
            "fp_layer1": fp_w1_stats,
            "ptq_layer1": ptq_w1_stats,
            "qat_layer1": qat_w1_stats,
        },
    }

    print(json.dumps(result, indent=2))
    print(f"\nSummary: FP={fp_accuracy*100:.2f}%, PTQ={ptq_accuracy*100:.2f}%, "
          f"QAT={qat_accuracy*100:.2f}%, delta(QAT-PTQ)={100*(qat_accuracy-ptq_accuracy):+.2f}%",
          file=sys.stderr)


if __name__ == "__main__":
    main()
