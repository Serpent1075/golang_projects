import Image from "next/image";
import Buttons from "./components/Buttons";


export default function Home() {
  
  return (
    <main className="flex min-h-screen flex-col items-center justify-between p-24">      
      <div className="">
        <p className="">
          Api test&nbsp;
          <code className="font-mono font-bold">src/app/page.tsx</code>
        </p>
        <Buttons />

      </div>
      
    </main>
  );
}
