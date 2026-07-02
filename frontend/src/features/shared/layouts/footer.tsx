import { Link } from '@tanstack/react-router';
import { Icon } from '../ui/icons';
import { Brand } from '../ui/brand';

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
                <li><Link to="/shop">Skincare</Link></li>
                <li><Link to="/shop">Haircare</Link></li>
                <li><Link to="/shop">Fragrance</Link></li>
                <li><Link to="/shop">Bodycare</Link></li>
                <li><Link to="/shop">Sets & Gifts</Link></li>
                <li><Link to="/shop">All products</Link></li>
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
                <li><a href="#">Shipping & delivery</a></li>
                <li><a href="#">Returns</a></li>
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
                  Community 18, Spintex<br />
                  <span>Adjacent KFC, Accra</span>
                </li>
                <li>
                  <Icon name="phone" size={14} />
                  0594 701 345
                </li>
                <li>
                  <Icon name="clock" size={14} />
                  Mon–Sat · 9am – 8pm
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div className="footer-bottom">
          <div>© 2026 Rue Cosmetics Ghana · All rights reserved</div>
          <div className="footer-legal">
            <a href="#">Privacy</a>
            <a href="#">Terms</a>
            <a href="#">Cookies</a>
          </div>
        </div>
      </div>
    </footer>
  );
}
