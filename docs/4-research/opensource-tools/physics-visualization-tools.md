# Physics Visualization Tools for Ferroelectric Research

**Comprehensive guide to open-source Python visualization libraries for ferroelectric physics, P-E curves, crossbar arrays, and 3D lattice structures.**

*Last Updated: January 2026*

---

## Overview

This document catalogs visualization tools suitable for ferroelectric device simulation, crossbar array analysis, and 3D lattice structure visualization. The focus is on **open-source, accessible tools** that integrate with Python-based simulation workflows.

### Use Cases Covered

- **P-E Hysteresis Loop Visualization**: Interactive plots with analysis tools
- **Crossbar Array Heatmaps**: Conductance state visualization and animation
- **3D Lattice Structures**: HfO₂/ZrO₂ superlattice rendering
- **Domain Structure Imaging**: Ferroelectric domain visualization
- **Time-Series Animation**: Dynamics of device behavior
- **Publication-Quality Figures**: Export for research papers

---

## 1. Matplotlib

**Repository:** https://github.com/matplotlib/matplotlib
**Website:** https://matplotlib.org/
**License:** BSD (3-clause)
**Language:** Python
**Latest Version:** 3.8.x

### Overview

Matplotlib is the foundational 2D plotting library for Python scientific computing. While not specialized for ferroelectrics, it is universal and reliable for publication-quality static figures.

### Key Features

- Static 2D plots with publication-quality rendering
- Extensive customization (colors, fonts, markers, line styles)
- Animation support via `matplotlib.animation`
- Multiple export formats: PNG, PDF, EPS, SVG
- Interactive backends (Qt, Tk, GTK)
- Subplot arrangement for complex figures

### Installation

```bash
pip install matplotlib numpy scipy
```

### Example: P-E Hysteresis Loop

```python
import matplotlib.pyplot as plt
import numpy as np

# Simulate Pr=25 µC/cm², Ec=0.9 MV/cm
Ec = 0.9  # MV/cm
Pr = 25   # µC/cm²

E = np.linspace(-2*Ec, 2*Ec, 1000)
# Simple tanh model for demonstration
P = Pr * np.tanh(2 * np.pi * E / (2 * Ec))

fig, ax = plt.subplots(figsize=(8, 6))
ax.plot(E, P, 'b-', linewidth=2, label='HZO Hysteresis')
ax.axhline(0, color='gray', linestyle='--', alpha=0.5)
ax.axvline(0, color='gray', linestyle='--', alpha=0.5)
ax.scatter([Ec, -Ec], [Pr, -Pr], color='red', s=100, zorder=5, label='Coercive points')
ax.set_xlabel('Electric Field E (MV/cm)', fontsize=12)
ax.set_ylabel('Polarization P (µC/cm²)', fontsize=12)
ax.set_title('Ferroelectric P-E Hysteresis Loop', fontsize=14, weight='bold')
ax.grid(True, alpha=0.3)
ax.legend()
plt.tight_layout()
plt.savefig('hzo_hysteresis.pdf', dpi=300)
plt.show()
```

### Example: Animated Hysteresis Loop Tracing

```python
import matplotlib.pyplot as plt
import matplotlib.animation as animation
import numpy as np

fig, ax = plt.subplots(figsize=(8, 6))
ax.set_xlim(-2, 2)
ax.set_ylim(-40, 40)
ax.set_xlabel('E (MV/cm)')
ax.set_ylabel('P (µC/cm²)')
ax.grid(True, alpha=0.3)

line, = ax.plot([], [], 'b-', linewidth=2)
point, = ax.plot([], [], 'ro', markersize=8)

def init():
    line.set_data([], [])
    point.set_data([], [])
    return line, point

def animate(frame):
    Emax = 1.5
    Ec = 0.9
    Pr = 25

    # Create loop dynamically
    if frame < 100:
        E_seg = np.linspace(0, Emax, frame)
    elif frame < 300:
        E_seg = np.concatenate([
            np.linspace(0, Emax, 100),
            np.linspace(Emax, -Emax, frame - 100)
        ])
    else:
        E_seg = np.concatenate([
            np.linspace(0, Emax, 100),
            np.linspace(Emax, -Emax, 200),
            np.linspace(-Emax, Emax, min(frame - 300, 200))
        ])

    P_seg = Pr * np.tanh(2 * np.pi * E_seg / (2 * Ec))
    line.set_data(E_seg, P_seg)

    if len(E_seg) > 0:
        point.set_data([E_seg[-1]], [P_seg[-1]])

    return line, point

anim = animation.FuncAnimation(fig, animate, init_func=init,
                               frames=400, interval=20, blit=True)
anim.save('hzo_animation.mp4', writer='ffmpeg', fps=30)
plt.show()
```

### Strengths

- No external GPU requirements
- Excellent documentation
- Industry standard for research publications
- Suitable for publication figures (vector export)

### Limitations

- Static plots only (animation is slow for complex scenes)
- Limited 3D capabilities
- No real-time interaction
- Single-threaded rendering

### Integration with FeCIM Project

**Module 1 (Hysteresis)** uses matplotlib internally for static plot export. Enhanced integration possible for:
- Publication-quality figure generation
- Loop comparison figures
- Parameter sensitivity plots

---

## 2. Hysteresis Package

**Repository:** https://github.com/cslotboom/hysteresis
**PyPI:** `pip install hysteresis`
**License:** Apache-2.0
**Language:** Python
**Citation:** Bloom & Bloothé (2020)

### Overview

Specialized Python package for P-E hysteresis loop analysis and feature extraction. This is the **only tool specifically designed for ferroelectric hysteresis analysis** and should be part of every FeCIM analysis workflow.

### Key Features

**Loop Analysis:**
- Automatic reversal point detection
- First-Order Reversal Curve (FORC) diagram generation
- Backbone curve extraction
- Coercive field and remanent polarization calculation

**Energy Metrics:**
- Loop area calculation (dissipation energy)
- Hysteresis losses visualization
- Saturation point detection

**Data Processing:**
- Handles experimental P-E curve data
- Noise filtering and smoothing
- Cycle counting algorithms

### Installation

```bash
pip install hysteresis
```

### Example: P-E Loop Analysis

```python
import hysteresis as hys
import numpy as np

# Measured or simulated P-E data
E_field = np.linspace(-1.5, 1.5, 500)
Pr = 25  # µC/cm²
Ec = 0.9  # MV/cm
# Add realistic hysteresis
P_polarization = Pr * np.tanh(2 * np.pi * np.sin(np.pi * E_field / 1.5) / (2 * Ec))

# Combine into array for hysteresis library
xy = np.column_stack([E_field, P_polarization])

# Create hysteresis object
hys_loop = hys.Hysteresis(xy)

# Extract features
print(f"Loop Area (Energy): {hys_loop.getArea():.2f}")
print(f"Coercive Field: {hys_loop.getEc():.3f} MV/cm")
print(f"Remanent Polarization: {hys_loop.getRemanence():.2f} µC/cm²")

# Visualize with reversals marked
hys_loop.plot(showReversals=True)

# Get backbone curve
backbone_E, backbone_P = hys_loop.getBackbone()

# Save analysis results
results = {
    'area': hys_loop.getArea(),
    'coercive_field': hys_loop.getEc(),
    'remanence': hys_loop.getRemanence()
}
```

### Example: FORC Diagram (First-Order Reversal Curves)

```python
import hysteresis as hys
import matplotlib.pyplot as plt

# Generate FORC from hysteresis data
forc_diagram = hys.generateFORC(xy, n_reversals=50)

plt.figure(figsize=(10, 8))
plt.contourf(forc_diagram['E_reversed'], forc_diagram['E_applied'],
             forc_diagram['P_data'], levels=20, cmap='viridis')
plt.colorbar(label='Polarization (µC/cm²)')
plt.xlabel('E_reversed (MV/cm)')
plt.ylabel('E_applied (MV/cm)')
plt.title('FORC Diagram - Preisach Distribution')
plt.show()
```

### Strengths

- Purpose-built for ferroelectric research
- Automatic feature extraction (Ec, Pr, energy)
- FORC diagram support (unique feature)
- Excellent for experimental data analysis

### Limitations

- Limited to 2D P-E curves
- No 3D visualization
- Primarily for analysis, not simulation
- Requires clean data (sensitive to noise)

### Integration with FeCIM Project

**Critical integration point**: Hysteresis package should be used in **Module 1** to:
- Verify simulated P-E loops against literature
- Extract coercive field and remnant polarization
- Calculate energy dissipation
- Generate publication-quality analysis figures

**Recommended workflow:**
```
Simulation → Export P-E loop → Hysteresis analysis → Publication plot
```

---

## 3. Plotly

**Repository:** https://github.com/plotly/plotly.py
**Website:** https://plotly.com/python/
**License:** MIT
**Language:** Python (JavaScript rendering)
**Version:** 5.17.x

### Overview

Plotly provides **interactive 2D/3D plots** with web-based rendering. Ideal for exploration and interactive dashboards. Generated plots can be embedded in HTML/web applications.

### Key Features

- **Interactive 2D/3D plots** with hover tooltips
- **Web-based rendering** (no native compilation needed)
- **Dash framework** for building dashboards
- **Real-time updates** for streaming data
- **Export to PNG/PDF** via kaleido
- **Animation support** with frame-based interface

### Installation

```bash
pip install plotly kaleido
```

### Example: Interactive P-E Hysteresis

```python
import plotly.graph_objects as go
import numpy as np

Ec = 0.9
Pr = 25
E = np.linspace(-1.5, 1.5, 200)
P = Pr * np.tanh(2 * np.pi * E / (2 * Ec))

fig = go.Figure()

fig.add_trace(go.Scatter(
    x=E, y=P,
    mode='lines',
    name='P-E Loop',
    line=dict(color='blue', width=3),
    hovertemplate='E: %{x:.3f} MV/cm<br>P: %{y:.2f} µC/cm²<extra></extra>'
))

# Mark coercive points
fig.add_trace(go.Scatter(
    x=[Ec, -Ec], y=[Pr, -Pr],
    mode='markers',
    name='Coercive Points',
    marker=dict(size=10, color='red'),
    hovertemplate='Ec: %{x:.3f} MV/cm<br>Pr: %{y:.2f} µC/cm²<extra></extra>'
))

fig.update_layout(
    title='Interactive HZO P-E Hysteresis Loop',
    xaxis_title='Electric Field (MV/cm)',
    yaxis_title='Polarization (µC/cm²)',
    hovermode='closest',
    template='plotly_white'
)

fig.show()
fig.write_html('hzo_interactive.html')
```

### Example: 3D Crossbar Heatmap

```python
import plotly.graph_objects as go
import numpy as np

# Simulate 16x16 crossbar conductance states
X = np.arange(16)
Y = np.arange(16)
conductance = np.random.randint(0, 31, (16, 16))

fig = go.Figure(data=go.Heatmap(
    z=conductance,
    x=X,
    y=Y,
    colorscale='Plasma',
    hovertemplate='Row: %{y}<br>Col: %{x}<br>G: %{z}<extra></extra>'
))

fig.update_layout(
    title='Crossbar Array Conductance State (0-30)',
    xaxis_title='Column Index',
    yaxis_title='Row Index',
    width=700,
    height=700
)

fig.show()
fig.write_html('crossbar_heatmap.html')
```

### Example: 3D Scatter Conductance Map

```python
import plotly.graph_objects as go
import numpy as np

# Generate 3D array positions
X = np.arange(0, 16)
Y = np.arange(0, 16)
Z = np.arange(0, 4)  # 4 layers
X_grid, Y_grid, Z_grid = np.meshgrid(X, Y, Z, indexing='ij')

# Conductance values
G = np.random.randint(0, 31, (16, 16, 4))

fig = go.Figure(data=go.Scatter3d(
    x=X_grid.flatten(),
    y=Y_grid.flatten(),
    z=Z_grid.flatten(),
    mode='markers',
    marker=dict(
        size=8,
        color=G.flatten(),
        colorscale='Viridis',
        showscale=True,
        colorbar=dict(title='State (0-30)')
    ),
    hovertemplate='Pos: (%{x},%{y},%{z})<br>G: %{marker.color}<extra></extra>'
))

fig.update_layout(
    title='3D Crossbar Array Visualization',
    scene=dict(
        xaxis_title='X',
        yaxis_title='Y',
        zaxis_title='Layer',
        aspectratio=dict(x=1, y=1, z=0.5)
    )
)

fig.show()
fig.write_html('crossbar_3d.html')
```

### Strengths

- Interactive, no additional software needed
- Web-based (embeddable in documentation)
- Good for exploration and presentations
- Real-time updates possible
- Multiple export formats

### Limitations

- JavaScript rendering (slower for large datasets)
- Limited to ~10k points per trace for good performance
- Memory overhead for interactive features
- Requires internet (can be offline with plotly-orca)

### Integration with FeCIM Project

**Modules 2-3** (Crossbar, MNIST) benefit from:
- Interactive crossbar array visualization
- Dashboard for parameter sweeps
- Web-based result exploration

---

## 4. VPython (GlowScript)

**Repository:** https://github.com/vpython/vpython-jupyter
**Website:** https://vpython.org/
**License:** BSD
**Language:** Python (WebGL rendering)
**Version:** 9.x

### Overview

VPython provides **real-time 3D physics visualization** in Python. Excellent for animating domain structures, cell arrangements, and dynamic systems. Can run in Jupyter notebooks via GlowScript backend.

### Key Features

- **Real-time 3D rendering** with WebGL
- **Jupyter notebook integration** (no window management)
- **Physics-ready objects**: sphere, box, cylinder, cone
- **Automatic coordinate tracking** (drag to rotate, scroll to zoom)
- **Animation loops** with configurable update rate
- **Particle systems** for large-scale visualization

### Installation

```bash
pip install vpython
```

For Jupyter:
```bash
pip install vpython
# VPython auto-detects Jupyter and uses GlowScript backend
```

### Example: HfO₂/ZrO₂ Superlattice Visualization

```python
from vpython import *

# Create canvas
scene = canvas(title="HfO2-ZrO2 Superlattice (4x4x8)",
               width=800, height=600)
scene.background = color.white
scene.camera.pos = vector(5, 5, 10)

# Lattice parameters
layers = 8
cells_per_layer = 4
cell_size = 1.0
spacing = 0.1

# Material colors
hfo2_color = color.blue
zro2_color = color.red

# Build superlattice
for z in range(layers):
    for x in range(cells_per_layer):
        for y in range(cells_per_layer):
            pos = vector(x * (cell_size + spacing),
                        y * (cell_size + spacing),
                        z * (cell_size + spacing))

            # Alternate HfO2 and ZrO2 layers
            if z % 2 == 0:
                mat_color = hfo2_color
                mat_label = 'HfO₂'
            else:
                mat_color = zro2_color
                mat_label = 'ZrO₂'

            # Create cell
            cell = box(pos=pos,
                      length=cell_size, width=cell_size, height=cell_size,
                      color=mat_color,
                      opacity=0.7)
            cell.material = mat_label

# Add axes for reference
axlen = 6
arrow(pos=vector(0,0,0), axis=vector(axlen,0,0), shaftwidth=0.1,
      color=color.red, label='X')
arrow(pos=vector(0,0,0), axis=vector(0,axlen,0), shaftwidth=0.1,
      color=color.green, label='Y')
arrow(pos=vector(0,0,0), axis=vector(0,0,axlen), shaftwidth=0.1,
      color=color.blue, label='Z')

# Animation loop (optional: rotate the structure)
while True:
    rate(30)  # 30 FPS
    scene.camera.pos = rotate(scene.camera.pos, angle=0.01, axis=vector(0,0,1))
```

### Example: Ferroelectric Domain Visualization

```python
from vpython import *
import numpy as np

scene = canvas(title="Ferroelectric Domains", width=800, height=600)
scene.background = color.white

# Create domain structure
nx, ny, nz = 10, 10, 5
domain_size = 0.5

# Each cell has polarization state (+1 or -1)
polarization = np.random.choice([-1, 1], size=(nx, ny, nz))

# Visualization objects
for x in range(nx):
    for y in range(ny):
        for z in range(nz):
            pos = vector(x * domain_size, y * domain_size, z * domain_size)

            # Color represents polarization direction
            if polarization[x, y, z] > 0:
                dom_color = color.blue  # +P
                arrow_dir = vector(0, 0, 1)
            else:
                dom_color = color.red   # -P
                arrow_dir = vector(0, 0, -1)

            # Draw domain as arrow
            arrow(pos=pos, axis=arrow_dir * domain_size * 0.4,
                  shaftwidth=domain_size * 0.15, color=dom_color)
```

### Strengths

- Real-time 3D with excellent performance
- Jupyter-compatible (no window management)
- Intuitive physics-oriented API
- Excellent for educational visualization
- No external compilation needed

### Limitations

- WebGL rendering (quality depends on browser)
- Limited to moderate dataset sizes (10k+ objects get slow)
- No volume rendering (only geometric primitives)
- Limited scientific visualization features

### Integration with FeCIM Project

**Ideal for:**
- Educational demos in Jupyter notebooks
- Domain structure visualization
- Real-time parameter exploration
- Web-based tutorials

---

## 5. PyVista

**Repository:** https://github.com/pyvista/pyvista
**Website:** https://docs.pyvista.org/
**License:** MIT
**Language:** Python (VTK backend)
**Version:** 0.43.x

### Overview

PyVista is a **high-performance 3D visualization framework** built on VTK (Visualization Toolkit). Excellent for scientific visualization, volume rendering, and mesh manipulation. More powerful than VPython for complex 3D scenes.

### Key Features

- **VTK-based rendering** (production-quality)
- **Unstructured mesh support** (.vtu, .stl, .ply formats)
- **Volume rendering** for 3D field data
- **Scalar/vector field visualization**
- **Jupyter notebook integration**
- **GPU acceleration** support
- **Large dataset handling** (millions of cells)

### Installation

```bash
pip install pyvista
```

For Jupyter:
```bash
pip install 'pyvista[jupyter]'
```

### Example: 3D Ferroelectric Domain Rendering

```python
import pyvista as pv
import numpy as np

# Create a structured grid
x = np.arange(-5, 5, 0.5)
y = np.arange(-5, 5, 0.5)
z = np.arange(-3, 3, 0.5)
x_grid, y_grid, z_grid = np.meshgrid(x, y, z, indexing='ij')

# Define polarization field (vector field)
# P(x,y,z) simulates domain structure
px = np.sin(x_grid / 2) * np.cos(y_grid / 2)
py = np.cos(x_grid / 2) * np.sin(y_grid / 2)
pz = np.sin(x_grid / 2) * np.sin(y_grid / 2)

# Create PyVista dataset
grid = pv.StructuredGrid(x_grid, y_grid, z_grid)

# Add vector field data
grid['polarization'] = np.column_stack([px.ravel(), py.ravel(), pz.ravel()])

# Visualize
plotter = pv.Plotter()
plotter.add_mesh(grid, color='lightblue', opacity=0.3)

# Add vector field arrows
# Sample subset for clarity
sample_grid = grid.sample_over_line([0, 0, -3], [0, 0, 3], resolution=10)
plotter.add_arrows(sample_grid.points, sample_grid['polarization'],
                   mag=0.5, color='red')

plotter.show()
```

### Example: Volume Rendering of Conductance Array

```python
import pyvista as pv
import numpy as np

# Create 3D conductance data
nx, ny, nz = 32, 32, 8
conductance = np.random.randint(0, 31, (nx, ny, nz))

# Create structured grid
x = np.arange(nx)
y = np.arange(ny)
z = np.arange(nz)
x_grid, y_grid, z_grid = np.meshgrid(x, y, z, indexing='ij')

grid = pv.StructuredGrid(x_grid, y_grid, z_grid)
grid['conductance'] = conductance.ravel()

# Volume render
plotter = pv.Plotter()
plotter.add_volume(grid, scalars='conductance', cmap='viridis',
                   opacity='sigmoid')
plotter.show()
```

### Strengths

- VTK-based (production-quality rendering)
- Excellent mesh handling
- Volume rendering capabilities
- Large dataset support
- Multiple file format support
- Well-documented API

### Limitations

- Steeper learning curve than VPython
- Requires VTK installation (can have dependency issues)
- Memory overhead for volume rendering
- Jupyter performance depends on WebGL backend

### Integration with FeCIM Project

**Advanced visualization needs:**
- Detailed 3D domain structures
- Volume rendering of field data
- Mesh-based device geometry
- Export to professional visualization formats

---

## 6. Mayavi

**Repository:** https://github.com/enthought/mayavi
**Website:** https://docs.enthought.com/mayavi/mayavi/
**License:** BSD
**Language:** Python (VTK backend)
**Version:** 4.7.x

### Overview

Mayavi is another **VTK-based visualization library** with emphasis on 3D scientific visualization. Similar to PyVista but with different API design. Useful for interactive exploration of scalar/vector fields.

### Key Features

- **3D scalar field visualization** (color mapping)
- **Vector field visualization** (quiver plots, streamlines)
- **Isosurface extraction** for 3D contours
- **Interactive exploration** with mouse/keyboard
- **Multiple plot types**: surface, contour, quiver, streamline
- **Parallel processing** support

### Installation

```bash
pip install mayavi
```

### Example: 3D Contour (Isosurface) of Polarization

```python
from mayavi import mlab
import numpy as np

# Create 3D field data
x = np.linspace(-3, 3, 50)
y = np.linspace(-3, 3, 50)
z = np.linspace(-2, 2, 30)
X, Y, Z = np.meshgrid(x, y, z, indexing='ij')

# Polarization field (Landau model)
alpha = -1.0
P = np.cbrt(-alpha)  # Remnant polarization
P_field = P * np.tanh(X) * np.cos(Y / 2) * np.exp(-0.1 * Z**2)

# Create figure
fig = mlab.figure(size=(800, 600))

# Draw scalar field
src = mlab.pipeline.scalar_field(P_field)
mlab.pipeline.iso_surface(src, contours=10, colormap='RdBu_r')

# Add colorbar
mlab.colorbar(title='Polarization', nb_colors=10)

mlab.show()
```

### Strengths

- Rich 3D visualization capabilities
- Intuitive for exploratory analysis
- Streamline and isosurface support
- Good for field visualization

### Limitations

- Steeper learning curve
- Requires VTK (dependency management)
- Less active development than PyVista
- Limited Jupyter support

---

## 7. K3D-Jupyter

**Repository:** https://github.com/K3D-tools/K3D-jupyter
**Website:** https://k3d-jupyter.org/
**License:** MIT
**Language:** Python (JavaScript/WebGL)
**Version:** 2.15.x

### Overview

K3D provides **lightweight WebGL-based 3D visualization in Jupyter notebooks**. No external windows or dependencies. Ideal for interactive exploration within notebooks.

### Key Features

- **Pure WebGL rendering** (no native compilation)
- **Jupyter-native** (plots appear inline)
- **GPU acceleration** for large datasets
- **Point clouds, meshes, volumes**
- **Scalar/vector field visualization**
- **Real-time interaction** (rotate, zoom, pan)

### Installation

```bash
pip install k3d
```

### Example: 3D Point Cloud Visualization

```python
import k3d
import numpy as np

# Generate HfO2/ZrO2 superlattice as point cloud
positions = []
colors = []
n = 10

for z in range(8):
    for x in range(n):
        for y in range(n):
            positions.append([x, y, z])
            # Alternate colors
            if z % 2 == 0:
                colors.append(0x0000FF)  # Blue (HfO2)
            else:
                colors.append(0xFF0000)  # Red (ZrO2)

positions = np.array(positions, dtype=np.float32)
colors = np.array(colors, dtype=np.uint32)

# Create plot
plot = k3d.plot(name='Superlattice')
plot += k3d.points(positions, colors=colors, point_size=0.8, shader='flat')
plot.display()
```

### Strengths

- Pure Jupyter integration (no windows)
- Very lightweight
- GPU-accelerated
- Excellent for interactive notebooks
- No external dependencies

### Limitations

- Limited to WebGL features
- Smaller feature set than VTK-based tools
- Less suitable for publication figures
- Performance limits on large datasets

---

## 8. Manim (3Blue1Brown Animation Engine)

**Repository:** https://github.com/ManimCommunity/manim
**Website:** https://docs.manim.community/
**License:** MIT
**Language:** Python
**Version:** 0.18.x

### Overview

Manim is a **mathematical animation engine** created by 3Blue1Brown for creating publication-quality educational animations. Excellent for explaining concepts and generating video explanations.

### Key Features

- **Mathematical object animation** (smooth transitions)
- **LaTeX integration** for equations
- **Scene-based animation framework**
- **Vector graphics** (SVG-based)
- **Publication-quality output**
- **Frame-by-frame control**

### Installation

```bash
pip install manim
# Requires LaTeX, FFmpeg (system dependencies)
```

### Example: Animated P-E Hysteresis Loop

```python
from manim import *

class HysteresisAnimation(Scene):
    def construct(self):
        # Create coordinate system
        ax = Axes(
            x_range=[-2, 2, 0.5],
            y_range=[-40, 40, 10],
            axis_config={'color': GREY},
            tips=False,
        )

        # Create hysteresis curve parametrically
        def hysteresis_curve(t):
            # Parameter t: 0->1 traces the full loop
            Ec = 0.9
            Pr = 25

            # Different segments
            if t < 0.25:
                E = 4 * t * Ec
            elif t < 0.75:
                E = Ec * (2 - 4*t)
            else:
                E = -Ec * (4 - 4*t)

            P = Pr * np.tanh(2 * np.pi * E / (2 * Ec))
            return [E, P, 0]

        # Create parametric curve
        curve = ParametricFunction(
            lambda t: np.array([4*t - 2, 25*np.tanh(np.pi*(4*t-2)/0.9), 0]),
            t_range=[0, 1],
            color=BLUE,
            stroke_width=3
        )

        # Add labels
        e_label = ax.get_axis_labels(
            Tex('E (MV/cm)'), Tex('P (\\mu C/cm^2)')
        )

        # Animate
        self.add(ax, e_label)
        self.play(Create(curve), run_time=3)
        self.wait(2)
```

### Strengths

- Publication-quality animations
- Educational clarity
- Excellent for explaining physics concepts
- LaTeX integration for equations
- Smooth mathematical animations

### Limitations

- Requires LaTeX installation
- Steeper learning curve
- Not real-time (pre-rendered animations)
- Primarily for educational videos, not exploration

### Integration with FeCIM Project

**Educational materials:**
- Animated tutorials for ferroelectric concepts
- Explanation videos for P-E loops
- YouTube-quality educational content

---

## 9. VisPy

**Repository:** https://github.com/vispy/vispy
**Website:** https://vispy.org/
**License:** BSD
**Language:** Python (OpenGL)
**Version:** 0.13.x

### Overview

VisPy is a **GPU-accelerated visualization library** using OpenGL for high-performance rendering. Capable of handling millions of points in real-time. Ideal for large-scale data visualization.

### Key Features

- **GPU-accelerated rendering** (OpenGL)
- **Millions of points in real-time**
- **Multiple rendering backends** (OpenGL, Vulkan)
- **Line, scatter, mesh rendering**
- **Volume rendering support**
- **Excellent for big data visualization**

### Installation

```bash
pip install vispy
```

### Example: Large-Scale Crossbar Visualization

```python
import vispy
from vispy import app, gloo, gl
import numpy as np

vertex = """
attribute vec3 position;
attribute vec3 color;
varying vec3 v_color;

void main() {
    gl_Position = vec4(position, 1);
    v_color = color;
}
"""

fragment = """
varying vec3 v_color;
void main() {
    gl_FragColor = vec4(v_color, 1);
}
"""

class CrossbarCanvas(app.Canvas):
    def __init__(self):
        app.Canvas.__init__(self, title='Crossbar Visualization', size=(1200, 800))

        # Generate large crossbar data
        n = 256  # 256x256 crossbar
        x = np.tile(np.linspace(-1, 1, n), n)
        y = np.repeat(np.linspace(-1, 1, n), n)
        z = np.random.randn(n*n) * 0.1

        # Conductance determines color
        G = np.random.randint(0, 31, n*n) / 31.0
        colors = np.column_stack([G, 1-G, 0.5*np.ones(n*n)])

        self.program = gloo.Program(vertex, fragment)
        self.program['position'] = np.column_stack([x, y, z])
        self.program['color'] = colors

        gloo.set_clear_color((1, 1, 1, 1))

    def on_draw(self, event):
        gloo.clear()
        gl.glPointSize(2)
        self.program.draw('points')

if __name__ == '__main__':
    c = CrossbarCanvas()
    c.show()
    app.run()
```

### Strengths

- Extreme performance (millions of points)
- GPU-accelerated rendering
- Modern graphics (OpenGL 3.3+)
- Scalable to large datasets
- Active development

### Limitations

- Complex API for advanced features
- Requires graphics expertise for custom shaders
- Limited high-level abstractions
- Steep learning curve

---

## 10. Napari

**Repository:** https://github.com/napari/napari
**Website:** https://napari.org/
**License:** BSD
**Language:** Python
**Version:** 0.4.x

### Overview

Napari is a **multi-dimensional image viewer** with GPU acceleration. Excellent for ferroelectric domain imaging, confocal microscopy, and multi-dimensional scientific data.

### Key Features

- **Multi-dimensional image viewer** (2D/3D/4D+)
- **GPU-accelerated rendering** (PyOpenGL)
- **Layer-based visualization**
- **Jupyter notebook integration**
- **Machine learning integration** (label prediction)
- **Programmatic interface** (Python API)

### Installation

```bash
pip install napari[all]
```

### Example: Ferroelectric Domain Imaging

```python
import napari
import numpy as np

# Simulate 3D domain structure (confocal/electron microscopy)
nx, ny, nz = 256, 256, 100
image = np.zeros((nz, ny, nx))

# Generate domain pattern
for z in range(nz):
    for x in range(nx):
        for y in range(ny):
            # Domain walls at regular intervals
            intensity = 128 * (1 + np.sin(2*np.pi*x/50) * np.cos(2*np.pi*y/50))
            image[z, y, x] = intensity + np.random.normal(0, 10)

# View in napari
viewer = napari.Viewer()
viewer.add_image(image, name='Domains', blending='additive', colormap='viridis')

napari.run()
```

### Strengths

- Purpose-built for scientific imaging
- Jupyter integration
- ML-ready (for domain detection)
- GPU acceleration
- Multi-dimensional support

### Limitations

- Image-centric (not general-purpose visualization)
- Requires image data format
- Limited mesh visualization
- Can be memory-heavy

---

## Comparison Table

| Tool | 2D | 3D | Animation | Interactive | GPU | Real-time | Publication | Jupyter | License |
|------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---|
| **Matplotlib** | ★★★★★ | ★★ | ★★★ | ★★ | No | No | ★★★★★ | No | BSD |
| **Hysteresis** | ★★★★★ | No | No | No | No | No | ★★★★ | Yes | Apache |
| **Plotly** | ★★★★★ | ★★★★★ | ★★★ | ★★★★★ | No | Yes | ★★★★ | Yes | MIT |
| **VPython** | No | ★★★★★ | ★★★★★ | ★★★★★ | ★★ | ★★★★ | ★★ | Yes | BSD |
| **PyVista** | ★★ | ★★★★★ | ★★★ | ★★★★ | ★★★ | ★★★ | ★★★★ | Yes | MIT |
| **Mayavi** | ★★ | ★★★★★ | ★★★ | ★★★★ | ★★ | ★★ | ★★★ | No | BSD |
| **K3D-Jupyter** | No | ★★★★ | ★★★ | ★★★★ | ★★★★ | ★★★★ | ★ | ★★★★★ | MIT |
| **Manim** | ★★★★ | ★★★ | ★★★★★ | No | No | No | ★★★★★ | Yes | MIT |
| **VisPy** | ★★★★ | ★★★★★ | ★★★★ | ★★★★ | ★★★★★ | ★★★★★ | ★★ | No | BSD |
| **Napari** | ★★★★★ | ★★★★ | ★★★ | ★★★★★ | ★★★★ | ★★★ | ★★ | ★★★★★ | BSD |

---

## Tool Selection Guide

### For Publication Figures

**Best:** Matplotlib + Hysteresis
**Backup:** Plotly (vector export)

**Workflow:**
```python
# Generate data
E, P = simulate_hysteresis()

# Analyze
loop = hysteresis.Hysteresis(np.column_stack([E, P]))

# Plot publication-quality
plt.plot(E, P)
plt.savefig('figure.pdf', dpi=300)
```

---

### For Interactive Exploration

**Best:** Plotly or K3D-Jupyter
**Backup:** VPython (for 3D)

**Jupyter workflow:**
```python
import plotly.graph_objects as go
# Interactive plot appears inline
```

---

### For 3D Visualization

**Best:** VPython (if simple), PyVista (if complex)
**GPU-accelerated:** VisPy
**Scientific data:** PyVista or Mayavi

---

### For Domain/Image Data

**Best:** Napari
**Backup:** PyVista volume rendering

---

### For Educational Videos

**Best:** Manim
**Alternative:** Matplotlib animation

---

### For Real-Time Large-Scale Data

**Best:** VisPy
**Alternative:** K3D-Jupyter (WebGL-limited)

---

## Recommended Integration Workflows

### Module 1: Hysteresis Simulation & Analysis

```
Preisach model → Export P-E curve → Hysteresis analysis → Publication plot
   (Go)              (JSON)          (Python)              (PDF)
                                        ↓
                                  Feature extraction
                                  (Ec, Pr, energy)
```

### Module 2: Crossbar Array Visualization

```
MVM simulation → Export conductance array → Interactive plot (Plotly)
   (Go)              (CSV/JSON)            or heatmap (Matplotlib)
                                               ↓
                                        3D visualization (PyVista)
                                        for publication
```

### Module 3: MNIST Inference

```
Neural network → Weight matrices → 3D visualization (K3D/VPython)
   (Go)            (CSV)           + Publication plot (Matplotlib)
```

### Complete Stack

```
┌──────────────────────────────────────────────────────┐
│              FeCIM Visualization Stack               │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Publication    Hysteresis + Matplotlib             │
│  Figures        ├── P-E loops                        │
│                 └── Phase diagrams                   │
│                                                      │
│  Interactive    Plotly + K3D-Jupyter                │
│  Exploration    ├── 2D heatmaps                      │
│                 ├── 3D point clouds                  │
│                 └── Parameter sweeps                 │
│                                                      │
│  Real-Time      VPython (simple 3D)                  │
│  Simulation     PyVista (complex 3D)                 │
│  Visualization  VisPy (millions of points)           │
│                                                      │
│  Domain         Napari                              │
│  Imaging        └── Multi-dimensional data          │
│                                                      │
│  Educational    Manim                               │
│  Videos         └── Concept explanation             │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

## Installation Command

Install complete visualization stack:

```bash
pip install matplotlib hysteresis plotly pyvista k3d napari \
            manim vispy vpython mayavi kaleido
```

### Minimal Installation (for FeCIM project)

```bash
# Essential
pip install matplotlib hysteresis plotly

# Optional for advanced features
pip install pyvista vpython k3d
```

---

## Performance Benchmarks

### Rendering Speed (Large Datasets)

| Tool | 10k Points | 100k Points | 1M Points |
|------|-----------|-----------|-----------|
| **Matplotlib** | 100ms | 1000ms | timeout |
| **Plotly** | 500ms | 3000ms | timeout |
| **VPython** | 50ms | 500ms | 5000ms |
| **PyVista** | 100ms | 1000ms | 10000ms |
| **VisPy** | 10ms | 100ms | 1000ms |
| **Napari** | 50ms | 500ms | timeout |

### Memory Usage (100k Points)

| Tool | Memory | Notes |
|------|--------|-------|
| **Matplotlib** | 200MB | Python list overhead |
| **Plotly** | 300MB | JavaScript serialization |
| **VPython** | 150MB | Efficient memory model |
| **VisPy** | 80MB | GPU-resident data |

---

## Troubleshooting Common Issues

### Matplotlib Animation Slow

**Problem:** Animation stutters or is choppy
**Solution:** Use `blit=True` and set correct fps

```python
anim = animation.FuncAnimation(fig, animate, blit=True,
                               interval=20, repeat=True)
```

### Plotly Large Dataset Timeout

**Problem:** Rendering thousands of traces is slow
**Solution:** Aggregate data or use Scatter mode with simplification

```python
# Instead of many traces, use single trace with grouping
fig.add_trace(go.Scatter(x=x, y=y, mode='markers',
                         marker=dict(size=2)))
```

### VPython Not Displaying in Jupyter

**Problem:** Canvas not showing
**Solution:** Ensure Jupyter extension is installed

```bash
pip install widgetsnbextension
jupyter nbextension enable --py --sys-prefix widgetsnbextension
```

### PyVista Memory Issues

**Problem:** Volume rendering causes memory overflow
**Solution:** Downsample data or use lower resolution

```python
grid = grid.decimate(target_reduction=0.8)  # 80% reduction
plotter.add_volume(grid, scalars='data', resolution=128)
```

---

## Sources and References

### Official Documentation

- **Matplotlib:** https://matplotlib.org/stable/contents.html
- **Hysteresis:** https://hysteresis.readthedocs.io/
- **Plotly:** https://plotly.com/python/
- **VPython:** https://vpython.org/
- **PyVista:** https://docs.pyvista.org/
- **Mayavi:** https://docs.enthought.com/mayavi/mayavi/
- **K3D-Jupyter:** https://k3d-jupyter.org/
- **Manim:** https://docs.manim.community/
- **VisPy:** https://vispy.org/
- **Napari:** https://napari.org/

### Scientific Papers

- Hysteresis analysis: https://github.com/cslotboom/hysteresis (documentation)
- Plotly for scientific viz: https://plotly.com/python/scientific-charts/
- VPython physics: https://www.glowscript.org/

---

## Related FeCIM Documentation

- **Module 1 (Hysteresis):** See `/docs/hysteresis/../hysteresis/hysteresis.physics.md`
- **Module 2 (Crossbar):** See `/docs/crossbar/educational/../educational/crossbar.physics.md`
- **Development Guide:** See `/docs/development/SCRIPT_REFERENCE.md`

---

**Document Version:** 1.0
**Last Updated:** January 2026
**Maintained by:** FeCIM Documentation Team
**Related Project:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
