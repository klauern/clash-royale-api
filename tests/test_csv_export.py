#!/usr/bin/env python3
"""
Test script for CSV export functionality.
"""

import asyncio
import sys
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parent / "src"))

from clash_royale_api import ClashRoyaleAPI, CSVExporter


async def test_csv_export():
    """Test CSV export functionality."""
    print("Testing CSV Export functionality...")

    # Initialize API client (requires valid API token in .env)
    try:
        api = ClashRoyaleAPI()
        print("✓ API client initialized successfully")
    except Exception as e:
        print(f"✗ Failed to initialize API client: {e}")
        print("\nNote: Make sure you have CLASH_ROYALE_API_TOKEN set in .env")
        return False

    # Test player tag (use a real tag for actual testing)
    test_player_tag = "#ABC123"  # Replace with a real player tag for testing

    # Test CSV exporter initialization
    try:
        exporter = CSVExporter(api, output_dir="./test_csv_output")
        print("✓ CSV exporter initialized successfully")
    except Exception as e:
        print(f"✗ Failed to initialize CSV exporter: {e}")
        return False

    # Test card database export (doesn't require player tag)
    try:
        print("\nTesting card database export...")
        filepath = await exporter.export_card_database_csv()
        print(f"✓ Card database exported to: {filepath}")
        if filepath.exists():
            print(f"  File size: {filepath.stat().st_size} bytes")
    except Exception as e:
        print(f"✗ Failed to export card database: {e}")

    print("\nCSV Export functionality test complete!")
    print("To test with real data, update the test_player_tag with a valid player tag.")

    return True


if __name__ == "__main__":
    success = asyncio.run(test_csv_export())
    sys.exit(0 if success else 1)