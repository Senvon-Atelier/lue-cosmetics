import type { ReactNode } from 'react';
import { STORE_INFO } from './store-info';

export type LegalPage = {
  slug: string;
  navLabel: string;
  title: ReactNode;
  lastUpdated: string;
  lead: string;
  body: ReactNode;
};

function ContactCard() {
  return (
    <div className="contact-card">
      <div className="k">Get in touch</div>
      <div className="v">
        {STORE_INFO.addressLine1}
        <br />
        {STORE_INFO.addressLine2}
        <br />
        {STORE_INFO.phone} · {STORE_INFO.hours}
      </div>
    </div>
  );
}

export const LEGAL_PAGES: readonly LegalPage[] = [
  {
    slug: 'privacy',
    navLabel: 'Privacy',
    title: (
      <>
        Privacy <em>policy</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'This policy explains what Lue Cosmetics collects when you shop with us, how we use it, and the choices you have. It is written to match how this store actually works — nothing more.',
    body: (
      <>
        <h2>What we collect</h2>
        <p>
          When you create an account we store your <strong>name</strong> and{' '}
          <strong>email</strong>. When you place an order we store your{' '}
          <strong>delivery address</strong> and <strong>order history</strong>{' '}
          so you can track and reorder. We collect only what we need to fulfil
          your orders.
        </p>
        <h2>Payments</h2>
        <p>
          Payments are processed by <strong>Paystack</strong>. Your full card
          details are entered on Paystack&rsquo;s secure checkout — Lue never
          sees or stores them. We keep a payment reference so we can match a
          payment to your order.
        </p>
        <h2>Cookies</h2>
        <p>
          We use only essential session and cart cookies. See our{' '}
          <strong>Cookie policy</strong> for details. We do not run
          advertising or third-party tracking cookies.
        </p>
        <h2>Your choices</h2>
        <p>
          You can view and update your details any time from your account. To
          delete your account and associated data, contact us using the details
          below.
        </p>
        <ContactCard />
      </>
    ),
  },
  {
    slug: 'terms',
    navLabel: 'Terms',
    title: (
      <>
        Terms <em>of service</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'These terms govern your use of the Lue Cosmetics store and the orders you place with us. By placing an order you agree to them.',
    body: (
      <>
        <h2>Orders &amp; acceptance</h2>
        <p>
          Placing an order is an offer to buy. An order is accepted once payment
          is confirmed; until then we may decline or cancel it (for example, if
          an item is unavailable).
        </p>
        <h2>Pricing</h2>
        <p>
          All prices are shown in <strong>Ghana cedis (GHS)</strong> and include
          any applicable taxes. We take care to price accurately; if a clear
          pricing error occurs we will contact you before proceeding.
        </p>
        <h2>Payment</h2>
        <p>
          Payment is taken through <strong>Paystack</strong>. Your order is
          confirmed once Paystack reports a successful charge.
        </p>
        <h2>Availability</h2>
        <p>
          Stock is limited and can sell out. If we cannot fulfil an item after
          payment, we will refund it to your original payment method.
        </p>
        <h2>Liability</h2>
        <p>
          To the extent permitted by law, Lue Cosmetics is not liable for
          indirect or consequential loss arising from use of the store. Nothing
          here limits rights you have under Ghanaian consumer law.
        </p>
        <h2>Governing law</h2>
        <p>These terms are governed by the laws of <strong>Ghana</strong>.</p>
      </>
    ),
  },
  {
    slug: 'cookies',
    navLabel: 'Cookies',
    title: (
      <>
        Cookie <em>policy</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'Cookies are small files a site stores in your browser. We keep our use of them to the minimum needed to run the store.',
    body: (
      <>
        <div className="callout info">
          Lue uses <strong>essential cookies only</strong> — no advertising or
          cross-site tracking.
        </div>
        <h2>What we use</h2>
        <p>
          A <strong>session cookie</strong> keeps you signed in, and a{' '}
          <strong>cart cookie</strong> remembers the items in your bag between
          visits. These are required for the store to function.
        </p>
        <h2>What we don&rsquo;t use</h2>
        <p>
          We do not set third-party advertising, profiling, or cross-site
          tracking cookies.
        </p>
        <h2>Managing cookies</h2>
        <p>
          You can clear or block cookies in your browser settings. Blocking the
          essential cookies above will stop the cart and sign-in from working.
        </p>
      </>
    ),
  },
  {
    slug: 'shipping',
    navLabel: 'Shipping & delivery',
    title: (
      <>
        Shipping <em>&amp; delivery</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'How and where we deliver, what it costs, and how long it takes.',
    body: (
      <>
        <div className="callout">
          Delivery times are estimates from dispatch and can vary with courier
          and location.
        </div>
        <h2>Rates</h2>
        <p>
          A flat delivery fee of <strong>GHS 25</strong> applies to every order.
          Orders of <strong>GHS 500 or more ship free</strong>.
        </p>
        <h2>Coverage</h2>
        <p>
          We deliver to all <strong>16 regions of Ghana</strong>. You choose
          your region and enter your address at checkout.
        </p>
        <h2>Timeframes</h2>
        <p>
          Greater Accra typically arrives within{' '}
          <strong>1&ndash;3 working days</strong>; other regions typically take{' '}
          <strong>3&ndash;7 working days</strong> after dispatch.
        </p>
        <h2>Tracking</h2>
        <p>
          You can follow your order&rsquo;s status any time from{' '}
          <strong>your account</strong>.
        </p>
      </>
    ),
  },
  {
    slug: 'returns',
    navLabel: 'Returns & refunds',
    title: (
      <>
        Returns <em>&amp; refunds</em>
      </>
    ),
    lastUpdated: 'July 2026',
    lead: 'We want you to love your order. Because cosmetics are personal-care products, some hygiene limits apply to returns.',
    body: (
      <>
        <div className="callout warn">
          For safety and hygiene, <strong>opened or used</strong> cosmetics
          cannot be returned unless they arrived damaged or faulty.
        </div>
        <h2>Return window</h2>
        <p>
          You can return <strong>unopened, unused</strong> items in their
          original packaging within <strong>7 days</strong> of delivery.
        </p>
        <h2>How to start a return</h2>
        <p>
          Contact us with your order number and we&rsquo;ll guide you through
          the next steps.
        </p>
        <h2>Refunds</h2>
        <p>
          Approved refunds are issued through <strong>Paystack</strong> to your
          original payment method, usually within a few working days of us
          receiving the item.
        </p>
        <h2>Damaged or wrong items</h2>
        <p>
          If an item arrives damaged or we sent the wrong product, we cover it —
          contact us and we&rsquo;ll make it right at no cost to you.
        </p>
        <ContactCard />
      </>
    ),
  },
];

export function getLegalPage(slug: string): LegalPage | undefined {
  return LEGAL_PAGES.find((p) => p.slug === slug);
}
