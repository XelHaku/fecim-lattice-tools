# Open-Source Neural Network and Quantization Tools

**A Comprehensive Guide to Tools for Training, Quantizing, and Deploying Neural Networks on CIM Hardware**

*Last Updated: January 2026*

---

## Overview

This document catalogs open-source tools, libraries, and frameworks for neural network training, weight quantization, hardware-aware training, and deployment on compute-in-memory systems. It covers tools from academic research, industry, and the open-source community relevant to the FeCIM MNIST demo.

**Note:** References to 30 levels refer to the demo baseline (conference claim; pending peer review). Peer‑reviewed devices report 32–140 states.

---

## 1. Deep Learning Frameworks

### 1.1 PyTorch

**URL:** https://pytorch.org/

**Description:** The most widely used research deep learning framework. Used for training our MNIST weights.

**Features:**
- Dynamic computation graphs
- Excellent GPU support
- Native quantization support
- Large ecosystem (torchvision, etc.)

**Installation:**
```bash
pip install torch torchvision
```

**MNIST Training Example:**
```python
import torch
import torch.nn as nn
from torchvision import datasets, transforms

class SimpleMLP(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc1 = nn.Linear(784, 128)
        self.fc2 = nn.Linear(128, 10)

    def forward(self, x):
        x = x.view(-1, 784)
        x = torch.relu(self.fc1(x))
        x = self.fc2(x)
        return x

# Load MNIST
transform = transforms.Compose([
    transforms.ToTensor(),
    transforms.Normalize((0.1307,), (0.3081,))
])
train_dataset = datasets.MNIST('./data', train=True, download=True, transform=transform)
train_loader = torch.utils.data.DataLoader(train_dataset, batch_size=64, shuffle=True)

# Train
model = SimpleMLP()
optimizer = torch.optim.Adam(model.parameters(), lr=0.001)
criterion = nn.CrossEntropyLoss()

for epoch in range(10):
    for data, target in train_loader:
        optimizer.zero_grad()
        output = model(data)
        loss = criterion(output, target)
        loss.backward()
        optimizer.step()
```

**Relevance to FeCIM:** We train models in PyTorch, export weights to JSON, load in Go.

---

### 1.2 TensorFlow/Keras

**URL:** https://tensorflow.org/

**Description:** Google's deep learning framework with strong production deployment support.

**Features:**
- TensorFlow Lite for edge deployment
- Excellent quantization tools
- TensorFlow Model Optimization Toolkit

**Installation:**
```bash
pip install tensorflow
```

**MNIST Training Example:**
```python
import tensorflow as tf

model = tf.keras.Sequential([
    tf.keras.layers.Flatten(input_shape=(28, 28)),
    tf.keras.layers.Dense(128, activation='relu'),
    tf.keras.layers.Dense(10)
])

model.compile(optimizer='adam',
              loss=tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True),
              metrics=['accuracy'])

mnist = tf.keras.datasets.mnist
(x_train, y_train), (x_test, y_test) = mnist.load_data()
x_train, x_test = x_train / 255.0, x_test / 255.0

model.fit(x_train, y_train, epochs=10)
```

---

### 1.3 JAX/Flax

**URL:** https://github.com/google/jax

**Description:** Google's high-performance numerical computing library with automatic differentiation.

**Features:**
- JIT compilation
- Automatic vectorization
- Functional programming style

**Installation:**
```bash
pip install jax jaxlib flax
```

---

## 2. Quantization Libraries

### 2.1 PyTorch Quantization (Native)

**URL:** https://pytorch.org/docs/stable/quantization.html

**Description:** PyTorch's built-in quantization support.

**Features:**
- Post-training quantization (PTQ)
- Quantization-aware training (QAT)
- Dynamic and static quantization
- Custom quantization configs

**Post-Training Quantization Example:**
```python
import torch
import torch.quantization as quant

# Original model
model = SimpleMLP()
model.load_state_dict(torch.load('mnist_fp.pt'))

# Configure quantization
model.qconfig = quant.get_default_qconfig('fbgemm')
quant.prepare(model, inplace=True)

# Calibrate with sample data
with torch.no_grad():
    for data, _ in calibration_loader:
        model(data)

# Convert to quantized model
quant.convert(model, inplace=True)
```

**Quantization-Aware Training Example:**
```python
# Prepare model for QAT
model.qconfig = quant.get_default_qat_qconfig('fbgemm')
quant.prepare_qat(model, inplace=True)

# Train with fake quantization
model.train()
for epoch in range(epochs):
    for data, target in train_loader:
        optimizer.zero_grad()
        output = model(data)
        loss = criterion(output, target)
        loss.backward()
        optimizer.step()

# Convert to quantized
model.eval()
quant.convert(model, inplace=True)
```

---

### 2.2 Brevitas (Xilinx)

**URL:** https://github.com/Xilinx/brevitas

**Description:** PyTorch library for quantization-aware training with arbitrary bit-widths.

**Features:**
- Arbitrary precision (not just 8-bit)
- Custom quantization schemes
- Binary and ternary networks
- Export to ONNX/FINN

**Installation:**
```bash
pip install brevitas
```

**Example (5-bit Weights for FeCIM):**
```python
import brevitas.nn as qnn
from brevitas.quant import Int8WeightPerTensorFloat

class QuantizedMLP(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc1 = qnn.QuantLinear(
            784, 128,
            weight_bit_width=5,  # ~30 levels (demo baseline)
            bias=True
        )
        self.fc2 = qnn.QuantLinear(
            128, 10,
            weight_bit_width=5,
            bias=True
        )

    def forward(self, x):
        x = x.view(-1, 784)
        x = torch.relu(self.fc1(x))
        x = self.fc2(x)
        return x
```

**Relevance to FeCIM:** Brevitas can train at exactly 5-bit precision (32 levels ≈ 30-level demo baseline).

---

### 2.3 TensorFlow Model Optimization Toolkit

**URL:** https://www.tensorflow.org/model_optimization

**Description:** TensorFlow's quantization and pruning toolkit.

**Features:**
- Quantization-aware training
- Post-training quantization
- Pruning and clustering
- TFLite conversion

**Example:**
```python
import tensorflow_model_optimization as tfmot

# QAT
quantize_model = tfmot.quantization.keras.quantize_model
q_aware_model = quantize_model(model)
q_aware_model.compile(optimizer='adam', loss='sparse_categorical_crossentropy')
q_aware_model.fit(x_train, y_train, epochs=10)

# Convert to TFLite
converter = tf.lite.TFLiteConverter.from_keras_model(q_aware_model)
converter.optimizations = [tf.lite.Optimize.DEFAULT]
tflite_model = converter.convert()
```

---

### 2.4 NNCF (Intel/OpenVINO)

**URL:** https://github.com/openvinotoolkit/nncf

**Description:** Neural Network Compression Framework for PyTorch and TensorFlow.

**Features:**
- Quantization (uniform and mixed precision)
- Pruning (filter and weight)
- Knowledge distillation
- NAS (Neural Architecture Search)

**Installation:**
```bash
pip install nncf
```

**Example:**
```python
from nncf import NNCFConfig
from nncf.torch import create_compressed_model

config = NNCFConfig.from_json({
    "compression": {
        "algorithm": "quantization",
        "weights": {"bits": 5},  # 32 levels
        "activations": {"bits": 8}
    }
})

compressed_model, compression_ctrl = create_compressed_model(model, config)
# Train compressed_model normally
```

---

### 2.5 Custom Quantization (Our Implementation)

**Location:** `module3-mnist/pkg/core/quantize.go`

**Description:** Go implementation of symmetric uniform quantization.

**Features:**
- Configurable levels (2-30)
- Symmetric range [-wMax, +wMax]
- Real-time re-quantization

**Example:**
```go
// Quantize weights to N levels
func QuantizeWeights(fpWeights [][]float64, levels int) [][]float64 {
    // Find max absolute value
    wMax := findMaxAbs(fpWeights)

    // Quantize each weight
    for i := range fpWeights {
        for j := range fpWeights[i] {
            normalized := (fpWeights[i][j] + wMax) / (2.0 * wMax)
            bin := int(math.Round(normalized * float64(levels-1)))
            quantized[i][j] = -wMax + float64(bin) * levelStep
        }
    }
    return quantized
}
```

---

## 3. Hardware-Aware Training Libraries

### 3.1 IBM AIHWKIT

**URL:** https://github.com/IBM/aihwkit

**Description:** IBM's analog AI hardware simulation and training library.

**Features:**
- Analog tile models (resistive crossbar)
- Device noise and non-idealities
- Hardware-in-the-loop training
- PyTorch integration

**Installation:**
```bash
pip install aihwkit
```

**Example (Hardware-Aware Training):**
```python
from aihwkit.nn import AnalogLinear
from aihwkit.simulator.configs import SingleRPUConfig
from aihwkit.simulator.configs.devices import ConstantStepDevice

# Configure analog device (FeCIM-like)
rpu_config = SingleRPUConfig(device=ConstantStepDevice())
rpu_config.device.w_min = -1.0
rpu_config.device.w_max = 1.0
rpu_config.device.w_min_dtod = 0.05  # 5% device-to-device variation
rpu_config.device.dw_min_dtod = 0.1  # 10% update variation

# Replace nn.Linear with AnalogLinear
class AnalogMLP(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc1 = AnalogLinear(784, 128, bias=True, rpu_config=rpu_config)
        self.fc2 = AnalogLinear(128, 10, bias=True, rpu_config=rpu_config)

    def forward(self, x):
        x = x.view(-1, 784)
        x = torch.relu(self.fc1(x))
        return self.fc2(x)

# Training with analog noise
model = AnalogMLP()
optimizer = torch.optim.SGD(model.parameters(), lr=0.1)

for epoch in range(epochs):
    for data, target in train_loader:
        optimizer.zero_grad()
        output = model(data)  # Forward pass includes device noise
        loss = criterion(output, target)
        loss.backward()
        optimizer.step()
```

**Relevance to FeCIM:** AIHWKIT provides the most realistic analog training simulation.

---

### 3.2 MemTorch

**URL:** https://github.com/coreylammie/MemTorch

**Description:** Memristive neural network simulation framework.

**Features:**
- Memristor device models
- Crossbar tile mapping
- IR drop and sneak path
- PyTorch Module replacement

**Installation:**
```bash
pip install memtorch
```

**Example:**
```python
import memtorch
from memtorch.mn.Module import patch_model
from memtorch.bh.crossbar.Program import naive_program
from memtorch.map.Parameter import naive_map

# Patch existing model with memristive simulation
patched_model = patch_model(
    model,
    memristor_model=memtorch.bh.memristor.LinearIonDrift,
    tile_shape=(128, 128),
    ADC_precision=6,
    DAC_precision=6,
    programming_routine=naive_program,
    mapping_routine=naive_map
)

# Inference includes memristor non-idealities
output = patched_model(input_data)
```

---

### 3.3 CrossSim (Sandia)

**URL:** https://github.com/sandialabs/cross-sim

**Description:** Crossbar array simulator with comprehensive non-ideality models.

**Features:**
- Full MVM/VMM simulation
- IR drop, sneak paths, variation
- Neural network layer mapping
- Custom device models

**Installation:**
```bash
git clone https://github.com/sandialabs/cross-sim
pip install -e cross-sim
```

**Example:**
```python
from cross_sim import CrossbarArray
from cross_sim.devices import RRAMDevice

array = CrossbarArray(
    rows=128,
    cols=64,
    device=RRAMDevice(Gmin=1e-6, Gmax=100e-6, sigma=0.05)
)

array.program_weights(weight_matrix)
output = array.mvm(input_vector, ir_drop=True, sneak_path=True)
```

---

## 4. MNIST-Specific Resources

### 4.1 torchvision.datasets.MNIST

**URL:** https://pytorch.org/vision/stable/generated/torchvision.datasets.MNIST.html

**Description:** Official MNIST dataset loader for PyTorch.

**Example:**
```python
from torchvision import datasets, transforms

transform = transforms.Compose([
    transforms.ToTensor(),
    transforms.Normalize((0.1307,), (0.3081,))
])

train_data = datasets.MNIST('./data', train=True, download=True, transform=transform)
test_data = datasets.MNIST('./data', train=False, transform=transform)
```

---

### 4.2 tensorflow.keras.datasets.mnist

**URL:** https://www.tensorflow.org/api_docs/python/tf/keras/datasets/mnist

**Description:** Official MNIST dataset loader for TensorFlow/Keras.

**Example:**
```python
import tensorflow as tf

(x_train, y_train), (x_test, y_test) = tf.keras.datasets.mnist.load_data()
x_train, x_test = x_train / 255.0, x_test / 255.0  # Normalize
```

---

### 4.3 EMNIST (Extended MNIST)

**URL:** https://www.nist.gov/itl/products-and-services/emnist-dataset

**Description:** Extended MNIST with letters and more digit variations.

**Features:**
- Multiple splits (digits, letters, balanced)
- 814,255 characters total
- Same 28×28 format

**Example:**
```python
from torchvision.datasets import EMNIST

# Load letters
dataset = EMNIST('./data', split='letters', download=True)
```

---

## 5. Weight Export and Conversion Tools

### 5.1 ONNX

**URL:** https://onnx.ai/

**Description:** Open Neural Network Exchange format for model interoperability.

**Features:**
- Framework-agnostic model representation
- Quantized model support
- Wide tool support

**Export from PyTorch:**
```python
import torch.onnx

torch.onnx.export(
    model,
    torch.randn(1, 784),
    "mnist.onnx",
    input_names=['input'],
    output_names=['output'],
    dynamic_axes={'input': {0: 'batch_size'}}
)
```

---

### 5.2 Custom JSON Export (Our Approach)

**Location:** `module3-mnist/scripts/export_weights.py`

**Description:** Export PyTorch weights to JSON for Go consumption.

**Example:**
```python
import torch
import json

model = torch.load('mnist_model.pt')

weights = {
    'layer1_weights': model['fc1.weight'].numpy().tolist(),
    'layer2_weights': model['fc2.weight'].numpy().tolist(),
    'biases1': model['fc1.bias'].numpy().tolist(),
    'biases2': model['fc2.bias'].numpy().tolist(),
}

with open('mnist_weights.json', 'w') as f:
    json.dump(weights, f)
```

**Loading in Go:**
```go
type WeightsFile struct {
    Layer1Weights [][]float64 `json:"layer1_weights"`
    Layer2Weights [][]float64 `json:"layer2_weights"`
    Biases1       []float64   `json:"biases1"`
    Biases2       []float64   `json:"biases2"`
}

func LoadWeights(filename string) (*WeightsFile, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    var wf WeightsFile
    json.Unmarshal(data, &wf)
    return &wf, nil
}
```

---

### 5.3 NumPy NPZ Format

**Description:** Efficient binary format for weight storage.

**Export:**
```python
import numpy as np

np.savez('mnist_weights.npz',
    fc1_weight=model.fc1.weight.detach().numpy(),
    fc1_bias=model.fc1.bias.detach().numpy(),
    fc2_weight=model.fc2.weight.detach().numpy(),
    fc2_bias=model.fc2.bias.detach().numpy()
)
```

**Read in Go (using gonpy):**
```bash
go get github.com/sbinet/npyio
```

```go
import "github.com/sbinet/npyio"

f, _ := npyio.Open("mnist_weights.npz")
fc1Weight := f.Get("fc1_weight")
```

---

## 6. Visualization Tools

### 6.1 TensorBoard

**URL:** https://www.tensorflow.org/tensorboard

**Description:** Visualization toolkit for training metrics.

**Features:**
- Loss/accuracy curves
- Weight histograms
- Graph visualization
- Embedding projector

**With PyTorch:**
```python
from torch.utils.tensorboard import SummaryWriter

writer = SummaryWriter('runs/mnist_experiment')

for epoch in range(epochs):
    # ... training ...
    writer.add_scalar('Loss/train', loss, epoch)
    writer.add_histogram('layer1/weights', model.fc1.weight, epoch)
```

**Run:**
```bash
tensorboard --logdir=runs
```

---

### 6.2 Weights & Biases (wandb)

**URL:** https://wandb.ai/

**Description:** Experiment tracking and visualization platform.

**Features:**
- Automatic logging
- Hyperparameter sweeps
- Model versioning
- Team collaboration

**Installation:**
```bash
pip install wandb
```

**Example:**
```python
import wandb

wandb.init(project="mnist-fecim")
wandb.config.levels = 30
wandb.config.noise = 0.05

for epoch in range(epochs):
    # ... training ...
    wandb.log({"loss": loss, "accuracy": accuracy})
```

---

### 6.3 Our FeCIM Visualizer (This Project)

**Location:** `module3-mnist/pkg/gui/`

**Description:** Go/Fyne real-time visualization.

**Features:**
- Interactive digit drawing
- Real-time FP vs CIM comparison
- Parameter sliders
- Weight heatmaps
- Activation visualization

**Key Files:**
```
pkg/gui/
├── app.go         # Main application
├── dualmode.go    # Dual-mode inference UI
├── canvas.go      # Drawing canvas
├── metrics.go     # Accuracy metrics
├── activations.go # Layer visualizations
└── tour.go        # Educational tour
```

---

## 7. Benchmarking Tools

### 7.1 PyTorch Profiler

**URL:** https://pytorch.org/docs/stable/profiler.html

**Description:** Built-in PyTorch profiling tools.

**Example:**
```python
from torch.profiler import profile, ProfilerActivity

with profile(activities=[ProfilerActivity.CPU, ProfilerActivity.CUDA]) as prof:
    model(input_data)

print(prof.key_averages().table(sort_by="cpu_time_total"))
```

---

### 7.2 NeuroSim Benchmarking

**URL:** https://github.com/neurosim

**Description:** Benchmark framework for neuromorphic hardware.

**Features:**
- Energy/latency/area estimation
- Technology node scaling
- Full DNN simulation

---

### 7.3 MLPerf Tiny

**URL:** https://mlcommons.org/en/inference-tiny/

**Description:** Standard benchmark for edge AI inference.

**MNIST Tasks:**
- Keyword spotting
- Visual wake words
- Image classification
- Anomaly detection

---

## 8. Tool Comparison Matrix

| Tool | Framework | PTQ | QAT | Arbitrary Bits | CIM Simulation |
|------|-----------|-----|-----|----------------|----------------|
| PyTorch Native | PyTorch | ✅ | ✅ | ❌ (8-bit) | ❌ |
| Brevitas | PyTorch | ✅ | ✅ | ✅ | ❌ |
| TF MOT | TensorFlow | ✅ | ✅ | ❌ | ❌ |
| NNCF | Both | ✅ | ✅ | ✅ | ❌ |
| AIHWKIT | PyTorch | ❌ | ✅ | ✅ | ✅ |
| MemTorch | PyTorch | ✅ | ❌ | ✅ | ✅ |
| CrossSim | Python | ✅ | ❌ | ✅ | ✅ |
| FeCIM (Ours) | Go | ✅ | ❌ | ✅ | ✅ |

---

## 9. Recommended Workflow for FeCIM

### 9.1 Training Pipeline

```
┌─────────────────────────────────────────────────────────────────────┐
│  RECOMMENDED WORKFLOW                                                │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  1. Train in PyTorch (FP32)                                         │
│     └── torchvision.datasets.MNIST                                  │
│     └── SimpleMLP(784→128→10)                                       │
│     └── 10 epochs → 97% accuracy                                    │
│                                                                     │
│  2. Quantization-Aware Training (Optional)                          │
│     └── Brevitas (5-bit weights)                                    │
│     └── OR AIHWKIT (noise-aware)                                    │
│     └── 10 epochs → 95% accuracy (quantized)                        │
│                                                                     │
│  3. Export to JSON                                                  │
│     └── Custom script (export_weights.py)                           │
│     └── Weights as nested arrays                                    │
│                                                                     │
│  4. Load in Go                                                      │
│     └── network.go:LoadWeights()                                    │
│     └── Requantize to current slider setting                        │
│                                                                     │
│  5. Run in FeCIM Visualizer                                         │
│     └── Dual-mode inference                                         │
│     └── Real-time visualization                                     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 9.2 Quick Start Script

```bash
#!/bin/bash
# Train and export MNIST weights for FeCIM

# Step 1: Train model
cd module3-mnist
python train_mnist_proper.py --epochs 10 --output mnist_model.pt

# Step 2: Export weights
python export_weights.py --input mnist_model.pt --output weights.json

# Step 3: Run visualizer
cd ..
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
```

---

## 10. Installation Quick Reference

```bash
# Core frameworks
pip install torch torchvision tensorflow

# Quantization tools
pip install brevitas
pip install tensorflow-model-optimization
pip install nncf

# Hardware simulation
pip install aihwkit
pip install memtorch
git clone https://github.com/sandialabs/cross-sim && pip install -e cross-sim

# Visualization
pip install tensorboard wandb

# Our FeCIM Visualizer
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
```

---

## 11. Community Resources

### 11.1 Documentation

- **PyTorch Quantization:** https://pytorch.org/docs/stable/quantization.html
- **Brevitas Docs:** https://xilinx.github.io/brevitas/
- **AIHWKIT Tutorials:** https://aihwkit.readthedocs.io/

### 11.2 Conferences

- **NeurIPS:** Neural Information Processing Systems
- **ICML:** International Conference on Machine Learning
- **MLSys:** Machine Learning and Systems
- **ISSCC:** International Solid-State Circuits Conference

### 11.3 Key Papers

1. "Quantization and Training of Neural Networks for Efficient Inference" (Jacob et al., 2018)
2. "DoReFa-Net: Training Low Bitwidth CNNs" (Zhou et al., 2016)
3. "BinaryConnect: Training Deep Neural Networks with Binary Weights" (Courbariaux et al., 2015)

---

## 12. Summary

### Best Tools by Use Case

| Use Case | Recommended Tool | Reason |
|----------|------------------|--------|
| **Training MNIST** | PyTorch + torchvision | Simplest setup |
| **QAT with arbitrary bits** | Brevitas | 5-bit ≈ FeCIM 30-level demo baseline |
| **Hardware-aware training** | AIHWKIT | Most realistic noise model |
| **CIM inference simulation** | FeCIM (Ours) | Real-time visualization |
| **Crossbar validation** | CrossSim | Comprehensive non-idealities |
| **Production deployment** | TensorFlow Lite | Best edge support |

### The Open-Source Neural Network Stack for FeCIM

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │ FeCIM Viz │  │ NeuroSim  │  │   wandb   │               │
│  │ (This)    │  │ (Bench)   │  │   (Track) │               │
│  └───────────┘  └───────────┘  └───────────┘               │
├─────────────────────────────────────────────────────────────┤
│                     Hardware Simulation                      │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │ AIHWKIT   │  │ MemTorch  │  │ CrossSim  │               │
│  │ (IBM)     │  │ (Sydney)  │  │ (Sandia)  │               │
│  └───────────┘  └───────────┘  └───────────┘               │
├─────────────────────────────────────────────────────────────┤
│                     Quantization Layer                       │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │ Brevitas  │  │ TF MOT    │  │   NNCF    │               │
│  │ (Xilinx)  │  │ (Google)  │  │  (Intel)  │               │
│  └───────────┘  └───────────┘  └───────────┘               │
├─────────────────────────────────────────────────────────────┤
│                     Framework Layer                          │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐               │
│  │  PyTorch  │  │ TensorFlow│  │    JAX    │               │
│  │ (Meta)    │  │ (Google)  │  │ (Google)  │               │
│  └───────────┘  └───────────┘  └───────────┘               │
└─────────────────────────────────────────────────────────────┘
```

---

## Related Documentation

- [MNIST Demo](mnist.demo.md) - Demo walkthrough and technical details
- [MNIST ELI5](mnist.ELI5.md) - Simple explanations for beginners
- [MNIST Research](mnist.research.md) - Academic background and literature review
- [Module Improvements Plan](mnist-module-improvements-plan.md) - Roadmap

---

*This document is part of the FeCIM Visualizer project.*
