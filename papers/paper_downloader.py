#!/usr/bin/env python3
"""
IronLattice Paper Downloader
Advanced paper downloading with API support, parallel downloads, and metadata extraction.
"""

import os
import sys
import json
import time
import hashlib
import argparse
import urllib.request
import urllib.parse
import urllib.error
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor, as_completed
from dataclasses import dataclass, field, asdict
from typing import Optional
import re
import ssl

# Disable SSL verification for some sites (use with caution)
ssl_context = ssl.create_default_context()
ssl_context.check_hostname = False
ssl_context.verify_mode = ssl.CERT_NONE

SCRIPT_DIR = Path(__file__).parent
DOWNLOAD_DIR = SCRIPT_DIR / "downloaded"
METADATA_FILE = SCRIPT_DIR / "paper_metadata.json"

# User agent for requests
USER_AGENT = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

@dataclass
class Paper:
    """Represents a scientific paper."""
    title: str
    filename: str
    url: Optional[str] = None
    doi: Optional[str] = None
    arxiv_id: Optional[str] = None
    authors: list = field(default_factory=list)
    year: Optional[int] = None
    source: str = "unknown"
    category: str = "general"
    downloaded: bool = False
    local_path: Optional[str] = None
    notes: str = ""


# ============================================================================
# Paper Database
# ============================================================================

PAPERS = [
    # Priority 1: Core HfO2/ZrO2 Ferroelectric Papers
    Paper(
        title="First-principles predictions of HfO2-based ferroelectric",
        filename="first_principles_HfO2_ferroelectric_2024",
        arxiv_id="2401.05288",
        year=2024,
        source="arxiv",
        category="core_ferroelectric"
    ),
    Paper(
        title="Ferroelectric CIM Review",
        filename="ferroelectric_CIM_review_2023",
        arxiv_id="2307.09357",
        year=2023,
        source="arxiv",
        category="cim"
    ),
    Paper(
        title="FerroX: GPU Phase-Field for Ferroelectrics",
        filename="ferrox_gpu_phasefield_2022",
        arxiv_id="2210.15668",
        year=2022,
        source="arxiv",
        category="simulation"
    ),
    Paper(
        title="Multi-Level FeFET Crossbar for Neural Networks",
        filename="multilevel_fefet_crossbar_2023",
        doi="10.1038/s41467-023-42110-y",
        url="https://www.nature.com/articles/s41467-023-42110-y.pdf",
        year=2023,
        source="nature",
        category="neuromorphic"
    ),
    Paper(
        title="FeCap/FeFET CIM Elements",
        filename="fecap_fefet_cim_elements_2024",
        doi="10.1038/s41598-024-59298-8",
        url="https://www.nature.com/articles/s41598-024-59298-8.pdf",
        year=2024,
        source="nature",
        category="cim"
    ),
    Paper(
        title="Adaptive Control Epitaxial HfO2/ZrO2",
        filename="adaptive_control_epitaxial_hfo2_zro2_2025",
        doi="10.1038/s41467-025-61758-2",
        url="https://www.nature.com/articles/s41467-025-61758-2.pdf",
        year=2025,
        source="nature",
        category="core_ferroelectric"
    ),
    Paper(
        title="Dual-Bit FeFET Enhanced Storage",
        filename="dual_bit_fefet_enhanced_storage_2025",
        doi="10.1038/s44335-025-00030-8",
        url="https://www.nature.com/articles/s44335-025-00030-8.pdf",
        year=2025,
        source="nature",
        category="memory"
    ),

    # Priority 2: CIM & Neuromorphic
    Paper(
        title="HCiM: ADC-Less Hybrid CIM",
        filename="hcim_adcless_hybrid_cim_2024",
        arxiv_id="2403.13577",
        year=2024,
        source="arxiv",
        category="cim"
    ),
    Paper(
        title="COMPASS: CIM Compiler Framework",
        filename="compass_compiler_framework_2025",
        arxiv_id="2501.06780",
        year=2025,
        source="arxiv",
        category="compiler"
    ),
    Paper(
        title="Simple Packing Algorithm for NVM",
        filename="simple_packing_algorithm_nvm_2024",
        arxiv_id="2411.04814",
        year=2024,
        source="arxiv",
        category="cim"
    ),
    Paper(
        title="Negative Feedback Training for NVCIM DNN Accelerators",
        filename="negative_feedback_training_nvcim_2023",
        arxiv_id="2307.xxxxx",  # Placeholder - needs lookup
        year=2023,
        source="arxiv",
        category="training"
    ),

    # Priority 3: Crossbar Non-Idealities
    Paper(
        title="Sneak Path in Self-Rectifying Crossbar Arrays",
        filename="sneak_path_self_rectifying_arrays_2022",
        url="https://www.frontiersin.org/articles/10.3389/femat.2022.988785/pdf",
        doi="10.3389/femat.2022.988785",
        year=2022,
        source="frontiers",
        category="crossbar"
    ),
    Paper(
        title="Hardware-Software Co-design for Non-idealities",
        filename="hw_sw_codesign_nonidealities_2024",
        doi="10.1007/s11432-024-4240-x",
        year=2024,
        source="springer",
        category="crossbar"
    ),

    # Priority 4: ACS Publications
    Paper(
        title="Ferroelectric HfO2-ZrO2 Multilayers with Reduced Wake-Up",
        filename="ferroelectric_hfo2_zro2_reduced_wakeup_2024",
        doi="10.1021/acsomega.4c10603",
        year=2024,
        source="acs",
        category="core_ferroelectric"
    ),
    Paper(
        title="HfO2-ZrO2 for 2D MoS2 NC-Transistors",
        filename="hfo2_zro2_2d_mos2_nc_transistors_2024",
        doi="10.1021/acsanm.4c04974",
        year=2024,
        source="acs",
        category="transistor"
    ),
    Paper(
        title="Ferroelectric Capacitors Superlattice Fatigue Stability",
        filename="ferroelectric_capacitors_superlattice_fatigue_2024",
        doi="10.1021/acsami.3c15732",
        year=2024,
        source="acs",
        category="reliability"
    ),

    # Priority 5: ACM Digital Library
    Paper(
        title="Extreme Partial-Sum Quantization (2-3 bit ADC)",
        filename="extreme_partial_sum_quantization_2022",
        doi="10.1145/3528104",
        year=2022,
        source="acm",
        category="quantization"
    ),
    Paper(
        title="Dynamic Quantization Range Control",
        filename="dynamic_quantization_range_control_2022",
        doi="10.1145/3498328",
        year=2022,
        source="acm",
        category="quantization"
    ),
    Paper(
        title="Variation Tolerant Weight Mapping",
        filename="variation_tolerant_mapping_2023",
        doi="10.1145/3585518",
        year=2023,
        source="acm",
        category="mapping"
    ),

    # Priority 6: ChemRxiv / Tour Lab
    Paper(
        title="Flash In2Se3 for Neuromorphic Computing (Tour Lab)",
        filename="flash_in2se3_neuromorphic_tour_2024",
        url="https://chemrxiv.org/engage/chemrxiv/article-details/659ef4cee9ebbb4db9de84cb",
        year=2024,
        authors=["Shin", "Jang", "Choi", "Kim", "Eddy", "Scotland", "Martin", "Han", "Tour"],
        source="chemrxiv",
        category="tour_lab"
    ),

    # Priority 7: RSC Papers
    Paper(
        title="Self-Rectifying FTJ Synapse Superlattice",
        filename="self_rectifying_ftj_synapse_superlattice_2024",
        doi="10.1039/d4mh00519h",
        year=2024,
        source="rsc",
        category="synapse"
    ),
    Paper(
        title="2T0C-FeDRAM Multi-bit Retention",
        filename="2t0c_fedram_multibit_retention_2024",
        doi="10.1039/d4nr02393e",
        year=2024,
        source="rsc",
        category="memory"
    ),

    # Priority 8: Wiley/Advanced Science
    Paper(
        title="2D SnS2 Analog Synaptic FeFET (>7-bit, 10^7 cycles)",
        filename="2d_sns2_analog_synaptic_fefet_2024",
        doi="10.1002/advs.202308588",
        url="https://advanced.onlinelibrary.wiley.com/doi/10.1002/advs.202308588",
        year=2024,
        source="wiley",
        category="synapse"
    ),
    Paper(
        title="Capacitive Crossbar Arrays Sneak-Free Design",
        filename="capacitive_crossbar_sneak_free_2021",
        doi="10.1002/aisy.202100258",
        url="https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202100258",
        year=2021,
        source="wiley",
        category="crossbar"
    ),

    # Priority 9: AIP Publications
    Paper(
        title="Oxygen Vacancy Dynamics in HfO2-ZrO2 Superlattice",
        filename="oxygen_vacancy_dynamics_superlattice_2024",
        doi="10.1063/5.0223518",
        year=2024,
        source="aip",
        category="reliability"
    ),

    # Priority 10: IEEE Papers
    Paper(
        title="BEOL-Compatible Superlattice FEFET Analog Synapse",
        filename="beol_superlattice_fefet_synapse_2022",
        doi="10.1109/TED.2022.xxxxx",  # Placeholder
        year=2022,
        authors=["Shin", "Tour"],
        source="ieee",
        category="tour_lab"
    ),
    Paper(
        title="HfO2-ZrO2 Superlattice FeFET Improved Endurance",
        filename="hfo2_zro2_superlattice_fefet_endurance_2023",
        url="https://ui.adsabs.harvard.edu/abs/2023ITED...70.3979P/abstract",
        year=2023,
        source="ieee",
        category="reliability"
    ),

    # Foundational Papers
    Paper(
        title="Ferroelectricity in Hafnium Oxide Thin Films (Böscke 2011)",
        filename="ferroelectricity_hfo2_boscke_2011",
        doi="10.1063/1.3634052",
        year=2011,
        authors=["Böscke", "Müller", "Bräuhaus", "Schröder", "Böttger"],
        source="aip",
        category="foundational"
    ),
    Paper(
        title="Ferroelectricity in Doped HfO2 (Park 2015)",
        filename="ferroelectricity_doped_hfo2_park_2015",
        doi="10.1002/adma.201501310",
        year=2015,
        authors=["Park", "Lee", "Kim", "Hwang"],
        source="wiley",
        category="foundational"
    ),

    # IBM AIHWKit Training
    Paper(
        title="IBM AIHWKit Hardware-Aware Training",
        filename="ibm_aihwkit_hwa_training",
        url="https://aihwkit.readthedocs.io/en/stable/hwa_training.html",
        year=2024,
        source="documentation",
        category="training",
        notes="Documentation, not paper"
    ),

    # Phase-field / Preisach Models
    Paper(
        title="Review of Preisach Models for Hysteresis",
        filename="preisach_models_hysteresis_review_2023",
        year=2023,
        source="pubmed",
        category="simulation"
    ),
    Paper(
        title="Time-Dependent Ginzburg-Landau Equation Algorithms",
        filename="tdgl_algorithms_2024",
        year=2024,
        source="osti",
        category="simulation"
    ),
    Paper(
        title="Phase-field Model of Multiferroic Composites",
        filename="phasefield_multiferroic_psu_2010",
        year=2010,
        source="psu",
        category="simulation"
    ),

    # =========== NEW PAPERS - Iteration 5 ===========

    # Tour Lab Neuromorphic
    Paper(
        title="In2Se3 Synthesized by FWF for Neuromorphic Computing",
        filename="in2se3_fwf_neuromorphic_tour_2024",
        url="https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aelm.202400603",
        doi="10.1002/aelm.202400603",
        year=2024,
        authors=["Shin", "Jang", "Choi", "Kim", "Eddy", "Scotland", "Martin", "Han", "Tour"],
        source="wiley",
        category="tour_lab",
        notes="Flash-within-flash synthesis, 87% MNIST accuracy"
    ),

    # High-Accuracy MNIST CIM
    Paper(
        title="Ferroelectric Memristor Crossbar Arrays for Neuromorphic RC",
        filename="ferroelectric_memristor_rc_2025",
        url="https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963",
        year=2025,
        source="sciencedirect",
        category="neuromorphic",
        notes="98.78% MNIST accuracy"
    ),
    Paper(
        title="CMOS-Compatible Ferroelectric Synaptic Arrays for CNN",
        filename="cmos_ferroelectric_synaptic_cnn_2022",
        url="https://www.science.org/doi/full/10.1126/sciadv.abm8537",
        doi="10.1126/sciadv.abm8537",
        year=2022,
        source="science",
        category="cim"
    ),
    Paper(
        title="2D Ferroelectric Hybrid CIM Hardware",
        filename="2d_ferroelectric_hybrid_cim_2024",
        url="https://www.science.org/doi/10.1126/sciadv.adp0174",
        doi="10.1126/sciadv.adp0174",
        year=2024,
        source="science",
        category="cim"
    ),

    # FeFET Linearity & Symmetry
    Paper(
        title="High Linearity ITO FeFETs for Neuromorphic",
        filename="high_linearity_ito_fefet_2025",
        url="https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202500078",
        doi="10.1002/aelm.202500078",
        year=2025,
        source="wiley",
        category="synapse",
        notes="Linearity 0.45/0.73, asymmetry 0.89"
    ),
    Paper(
        title="BEOL Analog FeFET for Online DNN Training",
        filename="beol_analog_fefet_training_2023",
        url="https://advanced.onlinelibrary.wiley.com/doi/full/10.1002/aisy.202300391",
        doi="10.1002/aisy.202300391",
        year=2023,
        source="wiley",
        category="training"
    ),
    Paper(
        title="Van der Waals FeFET Synapses",
        filename="vdw_fefet_synapses_2023",
        url="https://www.sciencedirect.com/science/article/pii/S2709472323000072",
        year=2023,
        source="sciencedirect",
        category="synapse",
        notes="128 states, Gmax/Gmin>120"
    ),

    # Simulation Frameworks
    Paper(
        title="FerroX: GPU Phase-Field Framework",
        filename="ferrox_gpu_phasefield_2023",
        arxiv_id="2210.15668",
        doi="10.1016/j.cpc.2023.108757",
        year=2023,
        source="arxiv",
        category="simulation",
        notes="AMReX-based, 15x GPU speedup"
    ),
    Paper(
        title="IBM AIHWKit for Neural Network Training",
        filename="ibm_aihwkit_paper_2023",
        url="https://pubs.aip.org/aip/aml/article/1/4/041102/2923573",
        doi="10.1063/5.0168089",
        year=2023,
        source="aip",
        category="simulation"
    ),
    Paper(
        title="Physical Reality of Preisach Model",
        filename="physical_reality_preisach_2018",
        url="https://www.nature.com/articles/s41467-018-06717-w",
        doi="10.1038/s41467-018-06717-w",
        year=2018,
        source="nature",
        category="simulation"
    ),

    # Recent Reviews
    Paper(
        title="Ferroelectric Devices for AI Chips Review",
        filename="ferroelectric_devices_ai_review_2025",
        url="https://www.sciencedirect.com/science/article/pii/S2709472325000036",
        year=2025,
        source="sciencedirect",
        category="review"
    ),
    Paper(
        title="HfO2 FeFET: Materials to Applications",
        filename="hfo2_fefet_review_2024",
        url="https://pubs.aip.org/aip/jap/article/138/1/010701/3351745",
        doi="10.1063/5.0216615",
        year=2024,
        source="aip",
        category="review"
    ),
    Paper(
        title="Recent Advances Ferroelectric Materials and CIM",
        filename="recent_advances_fe_cim_2025",
        url="https://link.springer.com/article/10.1186/s40580-025-00520-2",
        doi="10.1186/s40580-025-00520-2",
        year=2025,
        source="springer",
        category="review"
    ),
    Paper(
        title="Emerging 2D Ferroelectric for In-Sensor Computing",
        filename="emerging_2d_fe_insensor_2025",
        url="https://advanced.onlinelibrary.wiley.com/doi/10.1002/adma.202400332",
        doi="10.1002/adma.202400332",
        year=2025,
        source="wiley",
        category="review"
    ),

    # =========== NEW PAPERS - Iteration 6 (Physics & Visualization) ===========

    # Preisach Model Papers
    Paper(
        title="Transition-State-Theory Landau Double Well Ferroelectrics",
        filename="transition_state_landau_ferroelectric_2024",
        arxiv_id="2404.13138",
        year=2024,
        source="arxiv",
        category="simulation",
        notes="Connects Preisach, Landau, and NLS models"
    ),
    Paper(
        title="B-Spline Based Everett Map for Preisach Model",
        filename="bspline_everett_preisach_2024",
        arxiv_id="2410.02797",
        year=2024,
        source="arxiv",
        category="simulation",
        notes="B-spline Everett map construction for hysteresis"
    ),
    Paper(
        title="Newton Secant Methods for Preisach Control",
        filename="newton_secant_preisach_control_2024",
        arxiv_id="2406.11296",
        year=2024,
        source="arxiv",
        category="simulation",
        notes="Iterative remnant control of Preisach operators"
    ),

    # Phase-Field / TDGL Papers
    Paper(
        title="TopoTEM: Polarization Visualization STEM",
        filename="topotem_polarization_stem_2024",
        arxiv_id="2404.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="visualization",
        notes="Python TEMUL toolkit for polarization visualization"
    ),

    # Crossbar Simulation Papers
    Paper(
        title="Variation-Resilient FeFET In-Memory Probabilistic DL",
        filename="variation_resilient_fefet_bayesian_2023",
        arxiv_id="2312.xxxxx",  # Paper from Dec 2023
        year=2023,
        source="arxiv",
        category="neuromorphic",
        notes="Bayesian learning for FeFET variation mitigation"
    ),
    Paper(
        title="WAGONN Weight Bit Agglomeration Crossbar",
        filename="wagonn_crossbar_interconnect_2024",
        arxiv_id="2401.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="7nm FeFET crossbar interconnect resistance analysis"
    ),

    # FeFET Crossbar Design
    Paper(
        title="Ferroelectric Transistor Synaptic Crossbar Arrays",
        filename="fefet_synaptic_crossbar_device_circuit_2025",
        url="https://ieeexplore.ieee.org/document/xxxxx",  # IEEE 2024/2025
        year=2025,
        source="ieee",
        category="crossbar",
        notes="Device-circuit interactions, non-idealities on CiM accuracy"
    ),
    Paper(
        title="Ferroelectric Memristor HAO Structure Reservoir",
        filename="ferroelectric_memristor_hao_reservoir_2025",
        year=2025,
        source="nature",
        category="neuromorphic",
        notes="TiN/HAO/SiO2 structure, Pavlovian learning demo"
    ),

    # =========== NEW PAPERS - Iteration 7 (More arXiv) ===========

    # CIM Accuracy Benchmark Papers
    Paper(
        title="Memory Technologies Synaptic Crossbar Part 2 DNN Accuracy",
        filename="memory_tech_crossbar_dnn_accuracy_2024",
        arxiv_id="2408.05857",
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="FeFET vs SRAM vs ReRAM at 7nm, PWA accuracy boost 32.56%"
    ),
    Paper(
        title="Improving AIMC Accuracy Post-Training",
        filename="aimc_accuracy_post_training_2024",
        arxiv_id="2401.09859",
        year=2024,
        source="arxiv",
        category="cim",
        notes="Post-training optimization for analog in-memory computing"
    ),
    Paper(
        title="Pruning for ADC Efficiency Crossbar Accelerators",
        filename="pruning_adc_efficiency_crossbar_2024",
        arxiv_id="2403.13082",
        year=2024,
        source="arxiv",
        category="cim",
        notes="7.13x ADC energy improvement with <1% accuracy drop"
    ),
    Paper(
        title="CIM Landscape Research Commercial Overview",
        filename="cim_landscape_overview_2024",
        arxiv_id="2401.14428",
        year=2024,
        source="arxiv",
        category="review",
        notes="Comprehensive compute-near-memory and CIM overview"
    ),

    # Landau-Khalatnikov Simulation
    Paper(
        title="Landau-Khalatnikov Circuit Model Ferroelectric Hysteresis",
        filename="landau_khalatnikov_circuit_model_2001",
        arxiv_id="cond-mat/0108189",
        year=2001,
        source="arxiv",
        category="simulation",
        notes="Circuit equivalent of LK dynamical ferroelectric model"
    ),
    Paper(
        title="Atomistic Landau Model Ferroelectric MD Simulation",
        filename="atomistic_landau_ferroelectric_md_2022",
        arxiv_id="2206.12243",
        year=2022,
        source="arxiv",
        category="simulation",
        notes="Atomistic and coarse-grained molecular dynamics"
    ),

    # Domain Wall Dynamics Papers
    Paper(
        title="Dynamics Conducting Ferroelectric Domain Wall",
        filename="conducting_domain_wall_dynamics_2025",
        arxiv_id="2501.xxxxx",  # 2025 paper
        year=2025,
        source="arxiv",
        category="simulation",
        notes="TDGL + Schrodinger + Poisson for domain wall dynamics"
    ),
    Paper(
        title="Domain Wall Interfacial Ferroelectric Switching",
        filename="domain_wall_interfacial_switching_2025",
        arxiv_id="2502.xxxxx",  # 2025 paper
        year=2025,
        source="arxiv",
        category="simulation",
        notes="First-principles + ML for domain wall switching"
    ),
    Paper(
        title="Ferroelectric Domain Walls Environmental Sensors",
        filename="fe_domain_wall_sensors_2024",
        arxiv_id="2404.xxxxx",  # 2024 paper
        year=2024,
        source="arxiv",
        category="application",
        notes="Insulating to conducting domain wall states"
    ),

    # HfO2-ZrO2 Superlattice Papers
    Paper(
        title="Solution-Processed HfO2-ZrO2 Multilayer Ferroelectricity",
        filename="solution_hfo2_zro2_multilayer_2024",
        arxiv_id="2403.xxxxx",  # arXiv multilayer paper
        year=2024,
        source="arxiv",
        category="core_ferroelectric",
        notes="50nm thick multilayer, accelerated wake-up"
    ),
    Paper(
        title="First-Principles HfO2 Superlattice Design",
        filename="first_principles_hfo2_superlattice_2024",
        arxiv_id="2401.05288",  # Already have this one
        year=2024,
        source="arxiv",
        category="core_ferroelectric",
        notes="Polar/antipolar superlattice design"
    ),

    # FeFET Multibit Papers
    Paper(
        title="FeFET In-Memory Computing Multi-bit MAC",
        filename="fefet_cim_multibit_mac_2024",
        arxiv_id="2405.xxxxx",  # 2024 paper
        year=2024,
        source="arxiv",
        category="cim",
        notes="Multi-bit FeFET for MAC operations"
    ),
    Paper(
        title="FeFET Variability Aware Design Techniques",
        filename="fefet_variability_aware_2024",
        arxiv_id="2406.xxxxx",  # 2024 paper
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Variation-aware design for reliable FeFET computing"
    ),

    # =========== NEW PAPERS - Iteration 8 (Massive Expansion) ===========

    # Neuromorphic Spintronics
    Paper(
        title="Neuromorphic Spintronics Review",
        filename="neuromorphic_spintronics_review_2024",
        arxiv_id="2409.10290",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Comprehensive MRAM neuromorphic overview"
    ),
    Paper(
        title="2D Spintronics Neuromorphic Computing",
        filename="2d_spintronics_neuromorphic_2024",
        arxiv_id="2407.08469",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="MTJ skyrmion 2D spintronic devices"
    ),

    # Optical/Photonic Computing
    Paper(
        title="Optical Computing DNN Acceleration Survey",
        filename="optical_computing_dnn_survey_2024",
        arxiv_id="2407.21184",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Comprehensive photonic AI accelerator review"
    ),
    Paper(
        title="Photonic-Electronic Integrated AI Accelerators",
        filename="photonic_electronic_ai_accelerators_2024",
        arxiv_id="2403.14806",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Hybrid photonic-electronic accelerator review"
    ),
    Paper(
        title="Architecture-Level Photonic DNN Modeling",
        filename="photonic_dnn_architecture_modeling_2024",
        arxiv_id="2405.14806",
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="CiM tools for photonic system modeling"
    ),
    Paper(
        title="Photonic Neuromorphic CNN Accelerator",
        filename="photonic_neuromorphic_cnn_2024",
        arxiv_id="2405.08182",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Silicon photonic mesh, 98.6% MNIST accuracy"
    ),
    Paper(
        title="Mirage RNS Photonic DNN Training",
        filename="mirage_rns_photonic_training_2024",
        arxiv_id="2311.17323",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Residue Number System photonic training"
    ),

    # Transformer Hardware Accelerators
    Paper(
        title="Xpikeformer Hybrid Spiking Transformer",
        filename="xpikeformer_spiking_transformer_2024",
        arxiv_id="2407.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Analog-digital spiking transformer accelerator"
    ),
    Paper(
        title="AHWA-LoRA Analog Transformer Adaptation",
        filename="ahwa_lora_analog_transformer_2024",
        arxiv_id="2406.xxxxx",  # Need exact ID  
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Hardware-aware LoRA for AIMC transformers"
    ),
    Paper(
        title="Analog Attention Accelerator Charge CIM",
        filename="analog_attention_charge_cim_2024",
        arxiv_id="2405.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="Hybrid analog-digital attention with token pruning"
    ),
    Paper(
        title="AIMC Attention for Energy-Efficient LLMs",
        filename="aimc_attention_llm_2024",
        arxiv_id="2406.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Gain cell memory for LLM attention"
    ),

    # Analog Training Papers
    Paper(
        title="Exact Gradient Training on Analog Tiki-Taka",
        filename="tiki_taka_analog_training_2024",
        arxiv_id="2406.12774",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Theoretical foundation for analog gradient descent"
    ),
    Paper(
        title="Physical Neural Networks Analog Training",
        filename="physical_neural_networks_training_2024",
        arxiv_id="2412.xxxxx",  # Dec 2024
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="DSTD differentiable spike-time discretization"
    ),

    # ReRAM/RRAM Papers
    Paper(
        title="ReRAM Neural Field Reconstruction",
        filename="reram_neural_field_reconstruction_2024",
        arxiv_id="2407.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="40nm 256Kb ReRAM, 3D CT reconstruction"
    ),
    Paper(
        title="ReRAM Differential Equation Solver Diffusion",
        filename="reram_differential_diffusion_2024",
        arxiv_id="2406.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Analog diffusion model acceleration"
    ),
    Paper(
        title="ARAS Adaptive ReRAM DNN Accelerator",
        filename="aras_reram_dnn_accelerator_2024",
        arxiv_id="2403.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="Low-cost adaptive ReRAM accelerator"
    ),
    Paper(
        title="Cryogenic Analog 1T-ReRAM Cold NN",
        filename="cryogenic_reram_cold_nn_2024",
        arxiv_id="2401.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="77K operation, quantum computing compatible"
    ),
    Paper(
        title="ReRAM CIM SNN Reliability Survey",
        filename="reram_cim_snn_reliability_2024",
        arxiv_id="2403.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Device-level variation analysis for SNNs"
    ),

    # Ferroelectric Tunnel Junction (FTJ)
    Paper(
        title="Ferroelectric Materials Synaptic Transistors Review",
        filename="ferroelectric_synaptic_transistors_review_2024",
        arxiv_id="2406.13946",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="FTJ and FeFET synapse review"
    ),
    Paper(
        title="HfZrO FTJ Polarization Uniformity",
        filename="hfzro_ftj_polarization_2024",
        arxiv_id="2412.11288",
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="BEOL thermal budget FTJ optimization"
    ),

    # SNN Hardware
    Paper(
        title="SpikeExplorer FPGA SNN Hardware DSE",
        filename="spikeexplorer_fpga_snn_2024",
        arxiv_id="2403.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Python tool for SNN FPGA exploration"
    ),
    Paper(
        title="HASNAS Hardware-Aware SNN CIM Design",
        filename="hasnas_snn_cim_2024",
        arxiv_id="2403.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="SNN architecture search for CIM constraints"
    ),
    Paper(
        title="Intel Loihi RNN Adaptive Spiking",
        filename="loihi_rnn_adaptive_spiking_2024",
        arxiv_id="2401.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="3-bit weights, audio benchmark SOTA"
    ),

    # 3D Memory Accelerators
    Paper(
        title="Memory Neural Accelerators 3D",
        filename="memory_neural_accelerators_3d_2024",
        arxiv_id="2401.05363",
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="3D memory neural network acceleration"
    ),
    Paper(
        title="HBM Thermal Power Delivery LLM",
        filename="hbm_thermal_power_llm_2024",
        arxiv_id="2410.07524",
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="3D stacked HBM thermal optimization"
    ),
    Paper(
        title="CIM LLM Inference Overview",
        filename="cim_llm_inference_overview_2024",
        arxiv_id="2406.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="Compute-in-memory for LLM acceleration"
    ),
    Paper(
        title="3D Stacked Memory Side Acceleration",
        filename="3d_memory_side_acceleration_2024",
        arxiv_id="2403.04539",
        year=2024,
        source="arxiv",
        category="crossbar",
        notes="Processing-in-memory 3D design space"
    ),

    # HfO2 Switching and NCFET
    Paper(
        title="Interstitial Defects HfO2 Switching",
        filename="interstitial_defects_hfo2_switching_2024",
        arxiv_id="2407.xxxxx",  # July 2024
        year=2024,
        source="arxiv",
        category="core_ferroelectric",
        notes="Subnanosecond switching via defects"
    ),
    Paper(
        title="Neuromorphic Roadmap Emerging Technologies",
        filename="neuromorphic_roadmap_2024",
        arxiv_id="2407.xxxxx",  # July 2024 roadmap
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="PCM, MRAM, ReRAM neuromorphic roadmap"
    ),

    # Edge AI
    Paper(
        title="Spiking Transformer Edge 467x Energy",
        filename="spiking_transformer_edge_2024",
        arxiv_id="2403.xxxxx",  # Need exact ID
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="467x energy reduction vs edge GPU"
    ),

    # MRAM Neuromorphic
    Paper(
        title="Tokyo MRAM Binary Neural Network CIM",
        filename="mram_bnn_cim_tokyo_2024",
        arxiv_id="2411.xxxxx",  # Nov 2024
        year=2024,
        source="arxiv",
        category="neuromorphic",
        notes="Spintronic BNN training algorithm"
    ),

    # Domain Wall Dynamics (corrected)
    Paper(
        title="Conducting Ferroelectric Domain Wall Dynamics",
        filename="conducting_fe_domain_wall_2025",
        arxiv_id="2504.xxxxx",  # April 2025
        year=2025,
        source="arxiv",
        category="simulation",
        notes="TDGL + transport for domain wall dynamics"
    ),
]




def make_request(url: str, timeout: int = 30) -> Optional[bytes]:
    """Make HTTP request with proper headers."""
    headers = {
        "User-Agent": USER_AGENT,
        "Accept": "application/pdf,*/*",
    }
    req = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=timeout, context=ssl_context) as response:
            return response.read()
    except Exception as e:
        print(f"  Request failed: {e}")
        return None


def is_pdf(data: bytes) -> bool:
    """Check if data is a PDF file."""
    return data[:4] == b'%PDF'


def download_arxiv(paper: Paper) -> bool:
    """Download from arXiv."""
    if not paper.arxiv_id:
        return False

    url = f"https://arxiv.org/pdf/{paper.arxiv_id}.pdf"
    print(f"  Downloading arXiv: {paper.arxiv_id}")

    data = make_request(url, timeout=60)
    if data and is_pdf(data):
        output_path = DOWNLOAD_DIR / "arxiv" / f"{paper.filename}.pdf"
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_bytes(data)
        paper.local_path = str(output_path)
        paper.downloaded = True
        return True
    return False


def try_unpaywall(doi: str) -> Optional[str]:
    """Try to get open access URL from Unpaywall."""
    email = "ironlattice@example.com"
    url = f"https://api.unpaywall.org/v2/{doi}?email={email}"

    try:
        req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
        with urllib.request.urlopen(req, timeout=10) as response:
            data = json.loads(response.read().decode())
            if data.get("best_oa_location"):
                return data["best_oa_location"].get("url_for_pdf")
    except Exception:
        pass
    return None


def download_with_doi(paper: Paper) -> bool:
    """Try to download paper using DOI via Unpaywall."""
    if not paper.doi:
        return False

    print(f"  Checking Unpaywall for DOI: {paper.doi}")
    pdf_url = try_unpaywall(paper.doi)

    if pdf_url:
        print(f"  Found open access URL: {pdf_url}")
        data = make_request(pdf_url, timeout=60)
        if data and is_pdf(data):
            source_dir = paper.source if paper.source != "unknown" else "other"
            output_path = DOWNLOAD_DIR / source_dir / f"{paper.filename}.pdf"
            output_path.parent.mkdir(parents=True, exist_ok=True)
            output_path.write_bytes(data)
            paper.local_path = str(output_path)
            paper.downloaded = True
            return True
    return False


def download_direct(paper: Paper) -> bool:
    """Try direct URL download."""
    if not paper.url:
        return False

    print(f"  Trying direct URL: {paper.url}")

    # Try PDF URL variations
    urls_to_try = [paper.url]
    if not paper.url.endswith('.pdf'):
        urls_to_try.append(paper.url + '.pdf')
        urls_to_try.append(paper.url.replace('/abs/', '/pdf/'))

    for url in urls_to_try:
        data = make_request(url, timeout=60)
        if data and is_pdf(data):
            source_dir = paper.source if paper.source != "unknown" else "other"
            output_path = DOWNLOAD_DIR / source_dir / f"{paper.filename}.pdf"
            output_path.parent.mkdir(parents=True, exist_ok=True)
            output_path.write_bytes(data)
            paper.local_path = str(output_path)
            paper.downloaded = True
            return True

    return False


def download_frontiers(paper: Paper) -> bool:
    """Download from Frontiers (usually open access)."""
    if paper.source != "frontiers" or not paper.doi:
        return False

    # Frontiers PDFs
    article_id = paper.doi.split('/')[-1]
    url = f"https://www.frontiersin.org/articles/{paper.doi}/pdf"

    print(f"  Downloading Frontiers: {article_id}")
    data = make_request(url, timeout=60)
    if data and is_pdf(data):
        output_path = DOWNLOAD_DIR / "frontiers" / f"{paper.filename}.pdf"
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_bytes(data)
        paper.local_path = str(output_path)
        paper.downloaded = True
        return True
    return False


def download_paper(paper: Paper) -> bool:
    """Try all download methods for a paper."""
    print(f"\n[{paper.filename}] {paper.title}")

    # Check if already downloaded
    for source_dir in ["arxiv", "nature", "acs", "rsc", "aip", "frontiers",
                       "acm", "ieee", "springer", "wiley", "chemrxiv", "other"]:
        existing = DOWNLOAD_DIR / source_dir / f"{paper.filename}.pdf"
        if existing.exists():
            print(f"  Already downloaded: {existing}")
            paper.local_path = str(existing)
            paper.downloaded = True
            return True

    # Try different methods
    if paper.arxiv_id and download_arxiv(paper):
        print(f"  SUCCESS: Downloaded from arXiv")
        return True

    if paper.source == "frontiers" and download_frontiers(paper):
        print(f"  SUCCESS: Downloaded from Frontiers")
        return True

    if paper.url and download_direct(paper):
        print(f"  SUCCESS: Downloaded from direct URL")
        return True

    if paper.doi and download_with_doi(paper):
        print(f"  SUCCESS: Downloaded via Unpaywall")
        return True

    print(f"  FAILED: Could not download (may require subscription)")
    return False


# ============================================================================
# Semantic Scholar API
# ============================================================================

def search_semantic_scholar(query: str, limit: int = 10) -> list:
    """Search Semantic Scholar for papers."""
    encoded_query = urllib.parse.quote(query)
    url = f"https://api.semanticscholar.org/graph/v1/paper/search?query={encoded_query}&limit={limit}&fields=title,authors,year,openAccessPdf,externalIds,citationCount"

    try:
        req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
        with urllib.request.urlopen(req, timeout=15) as response:
            data = json.loads(response.read().decode())
            return data.get("data", [])
    except Exception as e:
        print(f"Search failed: {e}")
        return []


def print_search_results(results: list):
    """Print search results nicely."""
    for i, paper in enumerate(results, 1):
        title = paper.get("title", "Unknown")
        year = paper.get("year", "?")
        citations = paper.get("citationCount", 0)
        authors = ", ".join([a.get("name", "") for a in paper.get("authors", [])[:3]])
        if len(paper.get("authors", [])) > 3:
            authors += " et al."

        pdf_status = "PDF available" if paper.get("openAccessPdf") else "No open PDF"

        arxiv_id = paper.get("externalIds", {}).get("ArXiv", "")
        doi = paper.get("externalIds", {}).get("DOI", "")

        print(f"\n{i}. {title}")
        print(f"   Year: {year} | Citations: {citations} | {pdf_status}")
        print(f"   Authors: {authors}")
        if arxiv_id:
            print(f"   arXiv: {arxiv_id}")
        if doi:
            print(f"   DOI: {doi}")


# ============================================================================
# Main Functions
# ============================================================================

def download_all():
    """Download all papers in the database."""
    # Create directories
    for source in ["arxiv", "nature", "acs", "rsc", "aip", "frontiers",
                   "acm", "ieee", "springer", "wiley", "chemrxiv", "other"]:
        (DOWNLOAD_DIR / source).mkdir(parents=True, exist_ok=True)

    print("=" * 60)
    print("  IronLattice Paper Downloader")
    print("=" * 60)

    successful = 0
    failed = 0

    for paper in PAPERS:
        if download_paper(paper):
            successful += 1
        else:
            failed += 1
        time.sleep(0.5)  # Rate limiting

    # Save metadata
    metadata = [asdict(p) for p in PAPERS]
    METADATA_FILE.write_text(json.dumps(metadata, indent=2))

    print("\n" + "=" * 60)
    print("  Summary")
    print("=" * 60)
    print(f"  Downloaded: {successful}")
    print(f"  Failed: {failed}")
    print(f"  Metadata saved to: {METADATA_FILE}")
    print(f"  PDFs saved to: {DOWNLOAD_DIR}")


def search_papers(query: str, limit: int = 10):
    """Search for papers."""
    print(f"\nSearching Semantic Scholar: '{query}'")
    print("-" * 50)

    results = search_semantic_scholar(query, limit)
    if results:
        print_search_results(results)
    else:
        print("No results found.")


def search_tour_lab():
    """Search for external research group lab papers."""
    queries = [
        "external research group ferroelectric HfO2 ZrO2",
        "external research group neuromorphic computing",
        "Jaeho Shin superlattice FeFET",
        "external research institution ferroelectric memory",
    ]

    for query in queries:
        print(f"\n{'=' * 60}")
        search_papers(query, limit=5)


def list_papers():
    """List all papers in the database."""
    print("\n" + "=" * 60)
    print("  Paper Database")
    print("=" * 60)

    by_category = {}
    for paper in PAPERS:
        cat = paper.category
        if cat not in by_category:
            by_category[cat] = []
        by_category[cat].append(paper)

    for category, papers in sorted(by_category.items()):
        print(f"\n[{category.upper()}]")
        for paper in papers:
            status = "x" if paper.downloaded else " "
            print(f"  [{status}] {paper.title} ({paper.year or '?'})")
            if paper.doi:
                print(f"      DOI: {paper.doi}")
            if paper.arxiv_id:
                print(f"      arXiv: {paper.arxiv_id}")


def main():
    parser = argparse.ArgumentParser(
        description="IronLattice Paper Downloader",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s download          Download all papers
  %(prog)s search "query"    Search Semantic Scholar
  %(prog)s tour              Search for Tour lab papers
  %(prog)s list              List all papers in database
        """
    )

    subparsers = parser.add_subparsers(dest="command", help="Command to run")

    # Download command
    subparsers.add_parser("download", help="Download all papers")

    # Search command
    search_parser = subparsers.add_parser("search", help="Search for papers")
    search_parser.add_argument("query", help="Search query")
    search_parser.add_argument("-n", "--limit", type=int, default=10, help="Number of results")

    # Tour lab search
    subparsers.add_parser("tour", help="Search for external research group lab papers")

    # List papers
    subparsers.add_parser("list", help="List all papers in database")

    args = parser.parse_args()

    if args.command == "download":
        download_all()
    elif args.command == "search":
        search_papers(args.query, args.limit)
    elif args.command == "tour":
        search_tour_lab()
    elif args.command == "list":
        list_papers()
    else:
        parser.print_help()


if __name__ == "__main__":
    main()
