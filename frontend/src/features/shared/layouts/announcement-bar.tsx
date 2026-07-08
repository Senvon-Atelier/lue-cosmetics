// Announcement bar with marquee scrolling messages
// Ported from /rue/src/shared.jsx

const messages = [
  "Free delivery in Accra over GHS 250",
  "Community 18, Spintex — adjacent KFC",
  "Shop Mon–Sat · 9am–8pm",
  "New Lue Atelier fragrances have landed",
] as const;

export function AnnouncementBar() {
  return (
    <div className="announce">
      <div className="announce-track">
        {/* Duplicate messages for seamless loop */}
        {messages.map((msg, i) => (
          <span key={`first-${i}`}>
            {msg}
            <i />
          </span>
        ))}
        {messages.map((msg, i) => (
          <span key={`second-${i}`}>
            {msg}
            <i />
          </span>
        ))}
      </div>
    </div>
  );
}
