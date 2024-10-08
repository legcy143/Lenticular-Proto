"use client";
import { FileUpload } from "@/components/ui/file-upload";
import { RainbowButton } from "@/components/ui/rainbow-button";
import { useLenticularStore } from "@/store/useLenticularStore";
import React from "react";
import { Result } from "./_components/result-card";

function page() {
  const { generate, result, loading, images } = useLenticularStore();
  return (
    <div className="h-full w-full flex items-center justify-center  flex-col">
      {result ? (
        <Result />
      ) : (
        <>
          <FileUpload />
          <RainbowButton
            disabled={images.length === 0 || loading}
            onClick={() => generate()}
          >
            {loading ? "Generating..." : "Generate Lenticular Image"}
          </RainbowButton>
        </>
      )}
    </div>
  );
}

export default page;
