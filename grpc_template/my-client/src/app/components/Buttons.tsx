'use client';
import React from 'react'
import DarkButton from "@/components/DarkButton";
import { apiRequest } from "@/lib/utils";

const Buttons = () => {
    const onClickGPRC = (path: string) => {
        apiRequest(path);
      }    
  return (
    <div>
         <section className="flex flex-col top-4">
          <div className="flex flex-row top-4 items-center gap-4 text-[14px] mt-1">
            <DarkButton onClick={() => onClickGPRC("unary")} icon={<></>} className={"w-[230px] flex justify-center"} label={"unary grpc"} />
          </div>
          <div className="flex flex-row top-4 items-center gap-4 text-[14px] mt-1">
            <DarkButton onClick={() => onClickGPRC("server")} icon={<></>} className={"w-[230px] flex justify-center"} label={"server stream grpc"} />
          </div>
          <div className="flex flex-row top-4 items-center gap-4 text-[14px] mt-1">
            <DarkButton onClick={() => onClickGPRC("client")} icon={<></>} className={"w-[230px] flex justify-center"} label={"client stream grpc"} />
          </div>
          <div className="flex flex-row top-4 items-center gap-4 text-[14px] mt-1">
            <DarkButton onClick={() => onClickGPRC("bidirectional")} icon={<></>} className={"w-[230px] flex justify-center"} label={"bidirectional grpc"} />
          </div>
        </section>
    </div>
  )
}

export default Buttons