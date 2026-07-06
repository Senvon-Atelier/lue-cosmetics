import { useState } from 'react';

const testimonials = [
  {
    quote: 'The only shelf in Accra that carries everything I trust. I\'ve been a regular since they opened — the team knows my skin.',
    who: 'Ama Owusu',
    where: 'East Legon, Accra',
    since: 'Customer since 2023',
  },
  {
    quote: 'I ordered a gift for my sister in Kumasi. Arrived the next day, wrapped beautifully. They think about every part of it.',
    who: 'Kofi Mensah',
    where: 'Cantonments, Accra',
    since: 'Customer since 2024',
  },
  {
    quote: 'Walked in asking about retinol and left with a routine that finally worked. That\'s a rare kind of service.',
    who: 'Zainab Hassan',
    where: 'Tema, Ghana',
    since: 'Customer since 2022',
  },
];

export function TestimonialsSection() {
  const [idx, setIdx] = useState(0);

  const currentTestimonial = testimonials[idx];

  return (
    <section className="section testimonials">
      <div className="wrap testimonials-wrap">
        <div className="eyebrow" style={{ textAlign: 'center' }}>From our people</div>
        <blockquote className="quote">
          &ldquo;{currentTestimonial?.quote}&rdquo;
        </blockquote>
        <div className="quote-attrib">
          <div className="quote-name">{currentTestimonial?.who}</div>
          <div className="quote-meta">
            {currentTestimonial?.where} · {currentTestimonial?.since}
          </div>
        </div>
        <div className="quote-dots">
          {testimonials.map((_, i) => (
            <button
              key={i}
              className={`quote-dot ${i === idx ? 'active' : ''}`}
              onClick={() => setIdx(i)}
              aria-label={`View testimonial ${i + 1}`}
            />
          ))}
        </div>
      </div>
    </section>
  );
}
