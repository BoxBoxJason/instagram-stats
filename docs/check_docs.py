"""Fail when the documentation misses a configuration option or an output file.

The configuration options are the keys Config/Config.py reads from the configuration,
and the output files are named after the arguments __main__.py gives to the graph and
ranking functions. Run it from the repository root, with the standard library only:

    python3 -I docs/check_docs.py
"""
import ast
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parent.parent
PAGES = ROOT / "docs" / "content" / "docs"

# Output function name -> (index of the argument naming the file, file extension).
OUTPUT_FUNCTIONS = {
    "pieChartFigure": (2, ".png"),
    "heatMapFigure": (1, ".png"),
    "barGraphFigure": (2, ".png"),
    "textRanking": (4, ".txt"),
}


def calls(path):
    return [node for node in ast.walk(ast.parse(path.read_text(encoding="utf-8"))) if isinstance(node, ast.Call)]


def config_options():
    options = []
    for call in calls(ROOT / "Config" / "Config.py"):
        func = call.func
        if (isinstance(func, ast.Attribute) and func.attr == "get" and isinstance(func.value, ast.Name)
                and func.value.id == "configDict" and call.args and isinstance(call.args[0], ast.Constant)):
            options.append(call.args[0].value)
    return options


def output_files():
    files = []
    for call in calls(ROOT / "__main__.py"):
        if isinstance(call.func, ast.Name) and call.func.id in OUTPUT_FUNCTIONS:
            index, extension = OUTPUT_FUNCTIONS[call.func.id]
            files.append(call.args[index].value + extension)
    return files


def main():
    errors = []
    configuration = (PAGES / "configuration.md").read_text(encoding="utf-8")
    outputs = (PAGES / "outputs.md").read_text(encoding="utf-8")

    options = config_options()
    files = output_files()
    if not options or not files:
        errors.append("found no configuration option or no output file in the code, this check is out of date")

    errors += [f"configuration option {name!r} is not documented in docs/content/docs/configuration.md"
               for name in options if f"| `{name}` |" not in configuration]
    errors += [f"output file {name!r} is not documented in docs/content/docs/outputs.md"
               for name in files if f"| `{name}` |" not in outputs]

    for error in errors:
        print(error, file=sys.stderr)
    if errors:
        return 1
    print(f"{len(options)} configuration options and {len(files)} output files are documented")
    return 0


if __name__ == "__main__":
    sys.exit(main())
