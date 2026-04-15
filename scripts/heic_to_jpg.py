#!/usr/bin/env python3
"""
Converts a HEIC/HEIF file to JPEG using pillow-heif.
Usage: heic_to_jpg.py <input.heic> <output.jpg>
Exit 0 on success, non-zero on failure.
"""
import sys
import os

def main():
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <input.heic> <output.jpg>", file=sys.stderr)
        sys.exit(1)

    src, dst = sys.argv[1], sys.argv[2]

    try:
        import pillow_heif
        from PIL import Image
        pillow_heif.register_heif_opener()
        img = Image.open(src)
        # Convert to RGB if needed (e.g. RGBA, P modes aren't valid JPEG)
        if img.mode not in ("RGB", "L"):
            img = img.convert("RGB")
        img.save(dst, "JPEG", quality=90, optimize=True)
        print(f"Converted {src} -> {dst} ({img.size[0]}x{img.size[1]})")
        sys.exit(0)
    except ImportError as e:
        print(f"pillow_heif not available: {e}", file=sys.stderr)
        sys.exit(2)
    except Exception as e:
        print(f"Conversion failed: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
