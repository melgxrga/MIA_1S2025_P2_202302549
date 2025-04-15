'use client';
import { Editor } from '@monaco-editor/react';
import { useState, useEffect } from 'react';

const API_URL = 'http://localhost:8080';

export default function Home() {
  const [code, setCode] = useState<string>('');
  const [output, setOutput] = useState<string>('');

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    try {
      const res = await fetch(`${API_URL}/analyze`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command: code }),
      });

      await res.json();

      const consoleRes = await fetch(`${API_URL}/getConsole`);
      const consoleData = await consoleRes.json();

      setOutput(consoleData.console);
    } catch (error) {
      setOutput('Error en la ejecución');
    }
  };


  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file && file.name.endsWith('.smia')) {
      const reader = new FileReader();
      reader.onload = (e) => {
        setCode(e.target?.result as string);
      };
      reader.readAsText(file);
    } else {
      alert('Seleccione un archivo con extensión .smia');
    }
  };

  return (
    <div className='flex flex-col items-center justify-center min-h-screen py-2'>
      <h1 className='text-3xl font-bold mb-4'>Proyecto1 MIA</h1>
      <input type='file' accept='.glt' onChange={handleFileUpload} className='mb-4' />

      <div className='flex flex-row items-center justify-center w-full'>
        <div className='flex flex-col items-center justify-center w-1/2'>
          <Editor height='70vh' defaultLanguage='javascript' theme='vs-dark' value={code} onChange={(value) => setCode(value || '')} />
        </div>
        <div className='flex flex-col items-center justify-center w-1/2'>
          <Editor height='70vh' defaultLanguage='' theme='vs-dark' value={output} options={{ readOnly: true }} />
        </div>

      </div>

      <form onSubmit={handleSubmit} className='flex gap-4 mt-4'>
        <button type='submit' className='bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded'>
          Ejecutar
        </button>
      </form>
    </div>
  );
}
