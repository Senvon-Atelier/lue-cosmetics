import { Link } from '@tanstack/react-router';
import { Icon } from '../ui/icons';
import { Brand } from '../ui/brand';
import { STORE_INFO } from '../../../content/store-info';

export function Footer() {
  return (
    <footer className="footer">
      <div className="wrap">
        <div className="footer-top">
          <div className="footer-lead">
            <div className="footer-brand-logo">
              <Brand />
            </div>
            <p className="footer-blurb">
              Home of authentic beauty and wellness. A shelf of trusted names — and a few of our own — stocked in
              Accra, shipped across Ghana.
            </p>
            <div className="footer-socials">
              <a href="#" className="footer-social-link" aria-label="Instagram">
                <Icon name="instagram" size={18} />
              </a>
              <a href="#" className="footer-social-link" aria-label="TikTok">
                <Icon name="tiktok" size={18} />
              </a>
              <a href="#" className="footer-social-link" aria-label="WhatsApp">
                <Icon name="whatsapp" size={18} />
              </a>
            </div>
          </div>

          <div className="footer-cols">
            {/* Shop column */}
            <div className="footer-col">
              <h5>Shop</h5>
              <ul>
                <li><Link to="/shop" search={{}}>Skincare</Link></li>
                <li><Link to="/shop" search={{}}>Haircare</Link></li>
                <li><Link to="/shop" search={{}}>Fragrance</Link></li>
                <li><Link to="/shop" search={{}}>Bodycare</Link></li>
                <li><Link to="/shop" search={{}}>Sets & Gifts</Link></li>
                <li><Link to="/shop" search={{}}>All products</Link></li>
              </ul>
            </div>

            {/* Company column */}
            <div className="footer-col">
              <h5>Company</h5>
              <ul>
                <li><Link to="/about">About Rue</Link></li>
                <li><Link to="/" hash="journal">The Journal</Link></li>
                <li><a href="#">Store locator</a></li>
                <li><a href="#">Careers</a></li>
                <li><a href="#">Press</a></li>
              </ul>
            </div>

            {/* Help column */}
            <div className="footer-col">
              <h5>Help</h5>
              <ul>
                <li><a href="#">Contact us</a></li>
                <li><Link to="/legal/$slug" params={{ slug: 'shipping' }}>Shipping &amp; delivery</Link></li>
                <li><Link to="/legal/$slug" params={{ slug: 'returns' }}>Returns &amp; refunds</Link></li>
                <li><a href="#">FAQs</a></li>
                <li><a href="#">Authenticity</a></li>
              </ul>
            </div>

            {/* Visit column with icons */}
            <div className="footer-col">
              <h5>Visit the shop</h5>
              <ul className="footer-contact">
                <li>
                  <Icon name="pin" size={14} />
                  {STORE_INFO.addressLine1}<br />
                  <span>{STORE_INFO.addressLine2}</span>
                </li>
                <li>
                  <Icon name="phone" size={14} />
                  {STORE_INFO.phone}
                </li>
                <li>
                  <Icon name="clock" size={14} />
                  {STORE_INFO.hours}
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div className="footer-bottom">
          <div>© 2026 Rue Cosmetics Ghana · All rights reserved</div>
          <div className="footer-legal">
            <Link to="/legal/$slug" params={{ slug: 'privacy' }}>Privacy</Link>
            <Link to="/legal/$slug" params={{ slug: 'terms' }}>Terms</Link>
            <Link to="/legal/$slug" params={{ slug: 'cookies' }}>Cookies</Link>
          </div>
        </div>
      </div>
    </footer>
  );
}
