/*
	Copyright NetFoundry Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package templates

import (
	"bytes"
	"fmt"
	"strings"
	"ztna-core/ztna/logtrace"

	"github.com/russross/blackfriday"
)

const linebreak = "\n"

// ASCIIRenderer implements blackfriday.Renderer
var _ blackfriday.Renderer = &ASCIIRenderer{}

// ASCIIRenderer is a blackfriday.Renderer intended for rendering markdown
// documents as plain text, well suited for human reading on terminals.
type ASCIIRenderer struct {
	Indentation string

	listItemCount uint
	listLevel     uint
}

// NormalText gets a text chunk *after* the markdown syntax was already
// processed and does a final cleanup on things we don't expect here, like
// removing linebreaks on things that are not a paragraph break (auto unwrap).
func (r *ASCIIRenderer) NormalText(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	raw := string(text)
	lines := strings.Split(raw, linebreak)
	for _, line := range lines {
		trimmed := strings.Trim(line, " \n\t")
		out.WriteString(trimmed)
		out.WriteString(" ")
	}
}

// List renders the start and end of a list.
func (r *ASCIIRenderer) List(out *bytes.Buffer, text func() bool, flags int) {
	logtrace.LogWithFunctionName()
	r.listLevel++
	out.WriteString(linebreak)
	text()
	r.listLevel--
}

// ListItem renders list items and supports both ordered and unordered lists.
func (r *ASCIIRenderer) ListItem(out *bytes.Buffer, text []byte, flags int) {
	logtrace.LogWithFunctionName()
	if flags&blackfriday.LIST_ITEM_BEGINNING_OF_LIST != 0 {
		r.listItemCount = 1
	} else {
		r.listItemCount++
	}
	indent := strings.Repeat(r.Indentation, int(r.listLevel))
	var bullet string
	if flags&blackfriday.LIST_TYPE_ORDERED != 0 {
		bullet += fmt.Sprintf("%d.", r.listItemCount)
	} else {
		bullet += "*"
	}
	out.WriteString(indent + bullet + " ")
	r.fw(out, text)
	out.WriteString(linebreak)
}

// Paragraph renders the start and end of a paragraph.
func (r *ASCIIRenderer) Paragraph(out *bytes.Buffer, text func() bool) {
	logtrace.LogWithFunctionName()
	out.WriteString(linebreak)
	text()
	out.WriteString(linebreak)
}

// BlockCode renders a chunk of text that represents source code.
func (r *ASCIIRenderer) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	logtrace.LogWithFunctionName()
	out.WriteString(linebreak)
	lines := []string{}
	for _, line := range strings.Split(string(text), linebreak) {
		indented := r.Indentation + line
		lines = append(lines, indented)
	}
	out.WriteString(strings.Join(lines, linebreak))
}

func (r *ASCIIRenderer) GetFlags() int {
	logtrace.LogWithFunctionName()
	return 0
}
func (r *ASCIIRenderer) HRule(out *bytes.Buffer) {
	logtrace.LogWithFunctionName()
	out.WriteString(linebreak + "----------" + linebreak)
}
func (r *ASCIIRenderer) LineBreak(out *bytes.Buffer) {
	logtrace.LogWithFunctionName()
	out.WriteString(linebreak)
}
func (r *ASCIIRenderer) TitleBlock(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	logtrace.LogWithFunctionName()
	text()
}
func (r *ASCIIRenderer) BlockHtml(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) BlockQuote(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) TableRow(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) TableCell(out *bytes.Buffer, text []byte, align int) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) Footnotes(out *bytes.Buffer, text func() bool) {
	logtrace.LogWithFunctionName()
	text()
}
func (r *ASCIIRenderer) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	logtrace.LogWithFunctionName()
	r.fw(out, link)
}
func (r *ASCIIRenderer) CodeSpan(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) Emphasis(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) RawHtmlTag(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) TripleEmphasis(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) StrikeThrough(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	logtrace.LogWithFunctionName()
	r.fw(out, ref)
}
func (r *ASCIIRenderer) Entity(out *bytes.Buffer, entity []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, entity)
}
func (r *ASCIIRenderer) Smartypants(out *bytes.Buffer, text []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, text)
}
func (r *ASCIIRenderer) DocumentHeader(out *bytes.Buffer) {
	logtrace.LogWithFunctionName()
}
func (r *ASCIIRenderer) DocumentFooter(out *bytes.Buffer) {
	logtrace.LogWithFunctionName()
}
func (r *ASCIIRenderer) TocHeaderWithAnchor(text []byte, level int, anchor string) {
	logtrace.LogWithFunctionName()
}
func (r *ASCIIRenderer) TocHeader(text []byte, level int) {
	logtrace.LogWithFunctionName()
}
func (r *ASCIIRenderer) TocFinalize() {
	logtrace.LogWithFunctionName()
}

func (r *ASCIIRenderer) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	logtrace.LogWithFunctionName()
	r.fw(out, header, body)
}

func (r *ASCIIRenderer) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, link)
}

func (r *ASCIIRenderer) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	logtrace.LogWithFunctionName()
	r.fw(out, link)
}

func (r *ASCIIRenderer) fw(out *bytes.Buffer, text ...[]byte) {
	logtrace.LogWithFunctionName()
	for _, t := range text {
		out.Write(t)
	}
}
