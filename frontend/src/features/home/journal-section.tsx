import { Icon } from '../shared/ui/icons';

// TODO: Load journal images from API or CMS when content is dynamic.
const JOURNAL_IMAGE_MAP: Record<string, string> = {
  'b01': '/journal/journal-harmattan-ritual.jpg',
  'b02': '/journal/journal-slower-skincare.jpg',
  'b03': '/journal/journal-nuit-de-prelude.jpg',
};

const blogPosts = [
  {
    id: 'b01',
    tag: 'Rituals',
    readMin: 6,
    date: 'Mar 18',
    title: 'Building a Ghanaian Beauty Ritual for Harmattan Season',
    excerpt: 'When the dry wind rolls in, your skin asks for different things. Here\'s how our founder layers her routine from December to February.',
    tone: 'lavender' as const,
  },
  {
    id: 'b02',
    tag: 'Skincare',
    readMin: 9,
    date: 'Mar 04',
    title: 'The Case for Slower Skincare',
    excerpt: 'More steps isn\'t more care. A conversation with three estheticians in Accra about editing your shelf down to what actually works.',
    tone: 'cream' as const,
  },
  {
    id: 'b03',
    tag: 'Fragrance',
    readMin: 11,
    date: 'Feb 22',
    title: 'Inside The Lue Atelier: How Nuit de Prélude Came to Be',
    excerpt: 'Two years, seventeen drafts, and one perfumer in Grasse. The story behind our first in-house fragrance.',
    tone: 'ink' as const,
  },
];

export function JournalSection() {
  return (
    <section id="journal" className="section">
      <div className="wrap">
        <div className="section-head">
          <div>
            <div className="eyebrow">The Journal</div>
            <h2 className="h-display" style={{ fontSize: 'clamp(32px, 4vw, 56px)' }}>
              Read the <em>ritual.</em>
            </h2>
          </div>
          <a className="section-link">
            All stories <Icon name="arrow" size={14} />
          </a>
        </div>
        <div className="grid-3 journal-grid">
          {blogPosts.map((post) => (
            <a key={post.id} className="journal-card">
              <img src={JOURNAL_IMAGE_MAP[post.id]} alt={post.title} className="journal-card-img" />
              <div className="journal-body">
                <div className="journal-meta">
                  <span className="chip">{post.tag}</span>
                  <span>{post.readMin} min · {post.date}</span>
                </div>
                <h3 className="journal-title">{post.title}</h3>
                <p className="journal-excerpt">{post.excerpt}</p>
                <div className="journal-more">
                  Read story <Icon name="arrow" size={12} />
                </div>
              </div>
            </a>
          ))}
        </div>
      </div>
    </section>
  );
}
