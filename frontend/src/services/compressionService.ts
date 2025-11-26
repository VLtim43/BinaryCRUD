import {
  CompressFile,
  CompressAllFiles,
  DecompressFile,
  GetCompressedFiles,
  DeleteCompressedFile,
  GetBinFiles,
} from "../../wailsjs/go/main/App";

export interface CompressedFile {
  name: string;
  originalSize: number;
  compressedSize: number;
  algorithm: string;
  ratio: string;
  spaceSaved: string;
}

export interface CompressionResult {
  outputFile: string;
  originalSize: number;
  compressedSize: number;
  ratio: string;
  spaceSaved: string;
}

export interface BinFile {
  name: string;
  size: number;
}

export const compressionService = {
  compress: async (
    filename: string,
    algorithm: string
  ): Promise<CompressionResult> => {
    const result = await CompressFile(filename, algorithm);
    return {
      outputFile: result.outputFile as string,
      originalSize: result.originalSize as number,
      compressedSize: result.compressedSize as number,
      ratio: result.ratio as string,
      spaceSaved: result.spaceSaved as string,
    };
  },

  compressAll: async (algorithm: string): Promise<CompressionResult> => {
    const result = await CompressAllFiles(algorithm);
    return {
      outputFile: result.outputFile as string,
      originalSize: result.originalSize as number,
      compressedSize: result.compressedSize as number,
      ratio: result.ratio as string,
      spaceSaved: result.spaceSaved as string,
    };
  },

  decompress: async (filename: string): Promise<CompressionResult> => {
    const result = await DecompressFile(filename);
    return {
      outputFile: result.outputFile as string,
      originalSize: result.originalSize as number,
      compressedSize: result.compressedSize as number,
      ratio: result.ratio as string,
      spaceSaved: result.spaceSaved as string,
    };
  },

  getCompressedFiles: async (): Promise<CompressedFile[]> => {
    const files = await GetCompressedFiles();
    return files.map((f: any) => ({
      name: f.name as string,
      originalSize: f.originalSize as number,
      compressedSize: f.compressedSize as number,
      algorithm: f.algorithm as string,
      ratio: f.ratio as string,
      spaceSaved: f.spaceSaved as string,
    }));
  },

  deleteCompressedFile: async (filename: string): Promise<void> => {
    return DeleteCompressedFile(filename);
  },

  getBinFiles: async (): Promise<BinFile[]> => {
    const files = await GetBinFiles();
    return files.map((f: any) => ({
      name: f.name as string,
      size: f.size as number,
    }));
  },
};
