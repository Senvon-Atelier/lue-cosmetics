// Placeholder checkout API functions until backend endpoints are implemented
// These will be replaced by Orval-generated functions once the backend is ready

import axios from 'axios';

const API_BASE = '/api/v1';

export interface CheckoutInitRequest {
  address_id: string;
  shipping_method: string;
}

export interface CheckoutInitResponse {
  authorization_url: string;
  reference: string;
}

export interface CheckoutVerifyResponse {
  status: string;
  total_ghs_minor: number;
}

export async function postCheckoutInit(data: CheckoutInitRequest) {
  const response = await axios.post<CheckoutInitResponse>(`${API_BASE}/checkout/init`, data);
  return response;
}

export async function getCheckoutVerify(reference: string) {
  const response = await axios.get<CheckoutVerifyResponse>(`${API_BASE}/checkout/verify`, {
    params: { reference },
  });
  return response;
}
