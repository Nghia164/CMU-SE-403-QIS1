export interface UploadPhoto { id?: string | number; src?: string; alt?: string; /** placeholder glyph when no src */ icon?: string }
export interface PhotoUploaderProps {
  photos?: UploadPhoto[];
  max?: number;
  onAdd?: (files: File[]) => void;
  onRemove?: (id: string | number) => void;
  hint?: string;
}
export declare function PhotoUploader(props: PhotoUploaderProps): JSX.Element;
