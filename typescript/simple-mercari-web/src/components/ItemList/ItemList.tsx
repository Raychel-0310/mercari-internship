import React, { useEffect, useState } from 'react';

interface Item {
  id: number;
  name: string;
  category: string;
  image: string;
};

const server = process.env.REACT_APP_API_URL || 'http://127.0.0.1:9000';
const placeholderImage = process.env.PUBLIC_URL + '/logo192.png';

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
  const [items, setItems] = useState<Item[]>([])
  const fetchItems = () => {
    fetch(server.concat('/items'),
      {
        method: 'GET',
        mode: 'cors',
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json'
        },
      })
      .then(response => response.json())
      .then(data => {
        console.log('GET success:', data);
        setItems(data);
        // if (data && Array.isArray(data)) {
        //   setItems(data);
        // } else if (data && Array.isArray(data.items)) {
        //   setItems(data.items);
        // } else {
        //   console.error('Invalid data format:', data);
        //   setItems([]); 
        // }
        onLoadCompleted && onLoadCompleted();
      })
      .catch(error => {
        console.error('GET error:', error)
        setItems([]); 
      })
  }

  const getImgSrc = (image: string) => {
    return image ? `${server}/image/${image}` : placeholderImage;
  }

  useEffect(() => {
    if (reload) {
      fetchItems();
    }
  }, [reload]);

  return (
    <div className ="ItemList">
      
      {items.map((item) => {
        return (
          <div key={item.id} className='Item'>
            <img src={getImgSrc(item.image)} alt={item.name} />
            <p>
              <span>Name: {item.name}</span>
              <br />
              <span>Category: {item.category}</span>
            </p>
          </div>
        )
      })}
    </div>
  )
};
