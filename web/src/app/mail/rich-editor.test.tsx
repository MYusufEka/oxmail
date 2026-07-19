import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeAll, describe, it, expect, vi } from "vitest";
import { RichEditor } from "./rich-editor";

beforeAll(() => {
  document.elementFromPoint = () => document.body;
  Range.prototype.getClientRects = () => ({
    length: 0,
    item: () => null,
    [Symbol.iterator]: function* iterator() {},
  });
  Range.prototype.getBoundingClientRect = () => new DOMRect();
});

function getEditorElement() {
  const editorElement = document.querySelector(".tiptap");
  expect(editorElement).toBeInstanceOf(HTMLElement);
  return editorElement as HTMLElement;
}

describe("RichEditor", () => {
  it("renders initial HTML content", async () => {
    render(<RichEditor value="<p>Hello <strong>Oxmail</strong></p>" onChange={() => {}} />);

    await waitFor(() => {
      expect(getEditorElement()).toHaveTextContent("Hello Oxmail");
    });
  });

  it("calls onChange when content changes", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<RichEditor value="" onChange={onChange} />);

    const editorElement = getEditorElement();
    await user.click(editorElement);
    await user.keyboard("Hello");

    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith("<p>Hello</p>");
    });
  });

  it("applies bold formatting without calling legacy editing command", async () => {
    const commandName = ["exec", "Command"].join("");
    const legacyCommand = vi.fn();
    Object.defineProperty(document, commandName, {
      configurable: true,
      value: legacyCommand,
    });

    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<RichEditor value="" onChange={onChange} />);

    await user.click(getEditorElement());
    await user.click(screen.getByLabelText("Bold"));
    await user.keyboard("Hello");

    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith("<p><strong>Hello</strong></p>");
    });
    expect(legacyCommand).not.toHaveBeenCalled();
  });
});
