import { useState } from 'react';

// Mock testimonial data
const testimonials = [
  {
    quote: 'Rue has completely transformed my skincare routine. The products are authentic, the delivery is fast, and the concierge service helped me build a routine that actually works for my skin.',
    who: 'Amara O.',
    where: 'Accra',
    since: 'Customer since 2024',
  },
  {
    quote: 'Finally, a beauty store in Ghana that stocks the brands I actually trust. No more worrying about fake products or waiting weeks for international shipping. Rue is exactly what we needed.',
    who: 'Efia K.',
    where: 'Kumasi',
    since: 'Customer since 2025',
  },
  {
    quote: 'The WhatsApp concierge service is incredible. They took time to understand my skin concerns and recommended products within my budget. My skin has never looked better!',
    who: 'Nana A.',
    where: 'Takoradi',
    since: 'Customer since 2025',
  },
];

export function TestimonialsSection() {
  const [idx, setIdx] = useState(0);

  const currentTestimonial = testimonials[idx];

  return (
    <section className="section testimonials">
      <div className="wrap testimonials-wrap">
        <div className="eyebrow" style={{ textAlign: 'center' }}>From our people</div>
        <blockquote className="quote font-display">
          "{currentTestimonial?.quote}"
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
