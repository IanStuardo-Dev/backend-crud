import re
import unicodedata


TOKEN_ALIASES = {
    "café": "cafe",
    "cafes": "cafe",
    "coffee": "cafe",
    "coffe": "cafe",
    "instant": "instantaneo",
    "instantaneous": "instantaneo",
    "gr": "g",
    "grs": "g",
    "gramos": "g",
    "kilogramos": "kg",
    "kilos": "kg",
    "lts": "l",
    "litros": "l",
}


def strip_accents(text: str) -> str:
    decomposed = unicodedata.normalize("NFKD", text)
    return "".join(char for char in decomposed if not unicodedata.combining(char))


def canonicalize_token(token: str) -> str:
    return TOKEN_ALIASES.get(token, token)


def normalize_text(text: str) -> str:
    normalized = strip_accents(unicodedata.normalize("NFKC", text).lower())
    normalized = re.sub(r"(\d)([a-zA-Z]+)", r"\1 \2", normalized)
    normalized = re.sub(r"([a-zA-Z]+)(\d)", r"\1 \2", normalized)
    normalized = re.sub(r"[^a-z0-9\s]+", " ", normalized)
    normalized = re.sub(r"\s+", " ", normalized).strip()
    if not normalized:
        return ""

    tokens = [canonicalize_token(token) for token in normalized.split()]
    return " ".join(tokens)


def to_passage_text(text: str) -> str:
    normalized = normalize_text(text)
    if normalized.startswith("passage "):
        normalized = normalized[len("passage ") :]
    if normalized.startswith("passage:"):
        normalized = normalized[len("passage:") :]
    normalized = normalized.strip()
    return f"passage: {normalized}"
