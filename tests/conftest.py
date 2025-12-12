"""Pytest configuration and fixtures for Clash Royale API tests."""

import os
import pytest
from pathlib import Path


@pytest.fixture
def test_player_tag():
    """Return a test player tag for testing purposes."""
    return "#R8QGUQRCV"


@pytest.fixture
def mock_api_token():
    """Return a mock API token for testing."""
    return "test_api_token_12345"


@pytest.fixture
def temp_data_dir(tmp_path):
    """Create a temporary data directory for tests."""
    data_dir = tmp_path / "data"
    data_dir.mkdir()
    yield data_dir


@pytest.fixture(autouse=True)
def mock_env(monkeypatch, mock_api_token):
    """Mock environment variables for all tests."""
    monkeypatch.setenv("CLASH_ROYALE_API_TOKEN", mock_api_token)
    monkeypatch.setenv("DATA_DIR", "data")