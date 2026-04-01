import { Marked } from 'marked'
import DOMPurify from 'dompurify'

const md = new Marked({ breaks: true, gfm: true })

md.use({
  renderer: {
    link({ href, title, tokens }) {
      const text = this.parser.parseInline(tokens)
      const titleAttr = title ? ` title="${title}"` : ''
      return `<a href="${href}"${titleAttr} target="_blank" rel="noopener noreferrer">${text}</a>`
    },
  },
})

export function renderMarkdown(content: string): string {
  if (!content) return ''
  return DOMPurify.sanitize(md.parse(content) as string, {
    ADD_ATTR: ['target'],
  })
}
