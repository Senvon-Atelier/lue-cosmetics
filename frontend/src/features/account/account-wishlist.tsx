import { Link } from '@tanstack/react-router';
import { AcctHead } from './acct-primitives';

// Wishlist has no backend yet (tranche-1 spec §8.2) — honest empty state only.

export function AccountWishlist() {
  return (
    <main className="acct-main">
      <AcctHead eyebrow="Saved for later" title="Wishlist" />
      <div className="alert alert-info">
        Saving items isn't available yet — wishlist is coming soon.
      </div>
      <div className="acct-empty">
        <p>Nothing saved here yet. In the meantime, browse the edit.</p>
        <Link className="btn btn-primary" to="/shop" search={{}}>
          Explore products
        </Link>
      </div>
    </main>
  );
}
