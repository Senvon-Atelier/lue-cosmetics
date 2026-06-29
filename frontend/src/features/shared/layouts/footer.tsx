import { Link } from '@tanstack/react-router';
import { Icon } from '../ui/icons';
import { Brand } from '../ui/brand';

export function Footer() {
  return (
    <footer className="footer">
      <div className="wrap" style={{ maxWidth: 'var(--max)', margin: '0 auto', padding: '0 2rem' }}>
        <div className="footer-inner">
          <div className="footer-brand">
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
            <div className="footer-col">
              <h5>Shop</h5>
              <ul>
                <li>
                  <Link to="/shop">All Products</Link>
                </li>
                <li>
                  <Link to="/shop">Skincare</Link>
                </li>
                <li>
                  <Link to="/shop">Haircare</Link>
                </li>
                <li>
                  <Link to="/shop">Wellness</Link>
                </li>
              </ul>
            </div>
            <div className="footer-col">
              <h5>Company</h5>
              <ul>
                <li>
                  <Link to="/about">About Us</Link>
                </li>
                <li>
                  <Link to="/about">Our Story</Link>
                </li>
                <li>
                  <Link to="/about">Careers</Link>
                </li>
              </ul>
            </div>
            <div className="footer-col">
              <h5>Help</h5>
              <ul>
                <li>
                  <Link to="/account">My Account</Link>
                </li>
                <li>
                  <Link to="/account">Order Status</Link>
                </li>
                <li>
                  <Link to="/about">Contact Us</Link>
                </li>
              </ul>
            </div>
            <div className="footer-col">
              <h5>Visit</h5>
              <ul>
                <li>Spintex Road, Accra</li>
                <li>+233 20 123 4567</li>
                <li>Mon-Sat: 10am-7pm</li>
              </ul>
            </div>
          </div>
        </div>
        <div className="footer-bottom">
          <p>© 2026 Rue Cosmetics Ghana · All rights reserved</p>
          <div className="footer-legal">
            <a href="/legal/privacy">Privacy</a>
            <a href="/legal/terms">Terms</a>
            <a href="/legal/returns">Returns</a>
          </div>
        </div>
      </div>
    </footer>
  );
}
