import { NextRequest, NextResponse } from 'next/server';

export async function POST(req: NextRequest) {
  const { partition, username, password } = await req.json();

  // Llama al backend Go expuesto en localhost:8080/api/login
  try {
    const res = await fetch('http://localhost:8080/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: partition, user: username, pwd: password }),
    });
    const data = await res.json();
    return NextResponse.json(data, { status: res.status });
  } catch (e) {
    return NextResponse.json({ success: false, message: 'No se pudo conectar al backend.' }, { status: 500 });
  }
}
