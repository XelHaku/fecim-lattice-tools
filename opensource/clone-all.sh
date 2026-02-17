#!/bin/bash
# Clone all opensource tools referenced in docs/4-research/opensource-tools/
# Uses --depth 1 for shallow clones to save disk space
set -e

DEST="<local-path>"
cd "$DEST"

repos=(
  # Ferroelectric Simulation
  "https://github.com/mangerij/ferret"
  "https://github.com/AMReX-Microelectronics/FerroX"
  "https://github.com/materialsproject/pymatgen"
  "https://github.com/JAnderson419/Ferro"
  "https://github.com/supadupaplex/pfecap"
  "https://github.com/ferroelectrics/negativec"
  "https://github.com/idaholab/moose"
  "https://github.com/ramav87/FerroSim"
  "https://github.com/WMD-group/ferro_scripts"

  # Hysteresis Modeling
  "https://github.com/isaackramer/python-preisach"
  "https://github.com/fddf22/Preisachmodel"
  "https://github.com/cslotboom/hysteresis"
  "https://github.com/chiuczek/pyhist"
  "https://github.com/romanszewczyk/JAmodel"
  "https://github.com/IONICS-Lab/PEtra"

  # Circuit Simulation
  "https://github.com/Xyce/Xyce"
  "https://github.com/ra3xdh/qucs-s"
  "https://github.com/PySpice-org/PySpice"
  "https://github.com/ahkab/ahkab"
  "https://github.com/pascalkuthe/OpenVAF"

  # Circuit Analysis
  "https://github.com/coreylammie/MemTorch"
  "https://github.com/mph-/lcapy"
  "https://github.com/martok/py-symcircuit"
  "https://github.com/Ryan-O-Connor/pyjam"

  # Memristor / RRAM
  "https://github.com/technion-csl/vteam"
  "https://github.com/thomast8/Memristor-Models"
  "https://github.com/ZongxianYang0521/RRAM_model"
  "https://github.com/DUTh-FET/WaCPro"

  # Neural Network Hardware Mapping
  "https://github.com/Xilinx/brevitas"
  "https://github.com/IBM/aihwkit"
  "https://github.com/neurosim/DNN_NeuroSim_V2.1"
  "https://github.com/thu-nics/MNSIM-2.0"
  "https://github.com/google/qkeras"
  "https://github.com/openvinotoolkit/nncf"
  "https://github.com/Zhen-Dong/HAWQ"

  # EDA Tools
  "https://github.com/YosysHQ/yosys"
  "https://github.com/The-OpenROAD-Project/OpenROAD"
  "https://github.com/The-OpenROAD-Project/OpenLane"
  "https://github.com/The-OpenROAD-Project/OpenROAD-Tutorials"
  "https://github.com/RTimothyEdwards/magic"
  "https://github.com/RTimothyEdwards/netgen"
  "https://github.com/google/skywater-pdk"
  "https://github.com/google/gf180mcu-pdk"
  "https://github.com/IHP-GmbH/IHP-Open-PDK"
  "https://github.com/efabless/caravel"
  "https://github.com/efabless/open_source_chip_design"
  "https://github.com/mattvenn/caravel_layouts"
  "https://github.com/idea-fasoc/OpenFASOC"
  "https://github.com/ucb-art/BAG_framework"
  "https://github.com/iic-jku/IIC-OSIC-TOOLS"

  # Visualization
  "https://github.com/matplotlib/matplotlib"
  "https://github.com/plotly/plotly.py"
  "https://github.com/vpython/vpython-jupyter"
  "https://github.com/pyvista/pyvista"
  "https://github.com/enthought/mayavi"
  "https://github.com/K3D-tools/K3D-jupyter"
  "https://github.com/napari/napari"
  "https://github.com/vispy/vispy"
  "https://github.com/ManimCommunity/manim"

  # Scientific Computing
  "https://github.com/dealii/dealii"
  "https://github.com/usnistgov/atomman"
  "https://gitlab.com/QEF/q-e"

  # Data Acquisition
  "https://github.com/pyvisa/pyvisa"
  "https://github.com/pymeasure/pymeasure"
  "https://github.com/microsoft/Qcodes"
  "https://github.com/python-ivi/python-ivi"
  "https://github.com/ni/nidaqmx-python"
)

total=${#repos[@]}
cloned=0
skipped=0
failed=0

for url in "${repos[@]}"; do
  dirname=$(basename "$url" .git)
  if [ -d "$dirname" ]; then
    echo "SKIP ($dirname already exists)"
    skipped=$((skipped + 1))
  else
    echo "CLONING: $url -> $dirname"
    if git clone --depth 1 "$url" "$dirname" 2>&1; then
      cloned=$((cloned + 1))
    else
      echo "FAILED: $url"
      failed=$((failed + 1))
    fi
  fi
done

echo ""
echo "===== SUMMARY ====="
echo "Total repos: $total"
echo "Cloned: $cloned"
echo "Skipped (existed): $skipped"
echo "Failed: $failed"
