import React, { useState } from 'react';

const server = process.env.REACT_APP_API_URL || 'http://127.0.0.1:9000';

interface Prop {
  onListingCompleted?: () => void;
}

type formDataType = {
  name: string,
  category: string,
  image: string | File,
}

export const Listing: React.FC<Prop> = (props) => {
  const { onListingCompleted } = props;
  const initialState = {
    name: "",
    category: "",
    image: "",
  };
  const [values, setValues] = useState<formDataType>(initialState);
  const [preview, setPreview] = useState<string>("");

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values, [event.target.name]: event.target.value,
    })
  };
  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      const file = event.target.files[0];
      setValues({
        ...values, [event.target.name]: file,
      });
      const reader = new FileReader();
      reader.onloadend = () => {
        setPreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };
  const onSubmit = async(event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const data = new FormData()
    data.append('name', values.name)
    data.append('category', values.category)
    data.append('image', values.image)
    try {
      const response = await fetch(server.concat('/items'), {
        method: 'POST',
        body: data,
      });
  
      if (response.ok) {
        const responseData = await response.json();
  
        onListingCompleted && onListingCompleted(); 
      } else {
        console.error('エラー', response.statusText);
      }
    } catch (error) {
      console.error('POST error:', error);
    }
  };

  const displaySelectedFile = () => {
    return values.image instanceof File ? values.image.name : '選択されていません';
  };

  return (
    <div className='Listing'>
      <form onSubmit={onSubmit}>
        <div>
          <input type='text' name='name' id='name' placeholder='name' onChange={onValueChange} required />
          <input type='text' name='category' id='category' placeholder='category' onChange={onValueChange} />
          <input type='file' name='image' id='image' onChange={onFileChange} required />
          <button type='submit'>List this item</button>
          {preview && <img src={preview} alt="Preview" className="imagePreview" />}
        </div>
      </form>
    </div>
  );
}
