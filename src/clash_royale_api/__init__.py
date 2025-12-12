"""
Clash Royale API Data Collector

A comprehensive tool for collecting, analyzing, and tracking Clash Royale card data,
player statistics, and wild card information using the official Clash Royale API.
"""

from .api import ClashRoyaleAPI
from .csv_exporter import CSVExporter
from .deck_builder import DeckBuilder

__version__ = "0.1.0"
__author__ = "Your Name"
__email__ = "your.email@example.com"

__all__ = ["ClashRoyaleAPI", "CSVExporter", "DeckBuilder"]
