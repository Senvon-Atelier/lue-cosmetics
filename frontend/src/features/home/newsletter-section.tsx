import { Icon } from '../shared/ui/icons';

export function NewsletterSection() {
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    console.log('Newsletter signup submitted');
  };

  return (
    <section className="section nl">
      <div className="wrap nl-wrap">
        <div>
          <div className="eyebrow">Join the list</div>
          <h2 className="h-display" style={{ fontSize: 'clamp(32px, 5vw, 64px)' }}>
            Quiet emails.
            <br />
            <em>Good things</em> inside.
          </h2>
        </div>
        <form className="nl-form" onSubmit={handleSubmit}>
          <input type="email" placeholder="you@somewhere.com" />
          <button className="btn" type="submit" style={{ justifySelf: 'start', background: 'var(--lavender-300)', color: 'var(--ink)' }}>
            Subscribe <Icon name="arrow" size={14} />
          </button>
          <p className="nl-fine">One email a month. Unsubscribe whenever.</p>
        </form>
      </div>
    </section>
  );
}
