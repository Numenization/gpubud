import requests
import json
import datetime
import argparse
from bs4 import BeautifulSoup

def scrape(source):
    req = requests.get(source)
    soup = BeautifulSoup(req.content, 'html.parser')
    productGrid = soup.find('article', {'id': 'productGrid'}).find('ul')
    items = productGrid.find_all('li')

    data = {
        'timestamp': datetime.datetime.now().strftime('%m-%d-%Y %H:%M:%S'),
        'source': source,
        'gpus': []
    }
    for item in items:
        left = item.find('div', {'class': 'result_left'})
        right = item.find('div', {'class': 'result_right'})

        sku_p = item.find('p', {'class': 'sku'})

        link = left.find_next('a')
        link2 = link.find_next('a')
        count_span = right.find('span', {'class': 'inventoryCnt'})

        sku = sku_p.text.removeprefix('SKU: ')

        id = 0
        try:
            id = int(link['data-id'])
        except ValueError:
            pass

        stock = 0
        try:
            if count_span is not None:
                stock = int(count_span.contents[0].strip())
        except ValueError:
            pass

        price = 0
        try:
            price = float(link['data-price'])
        except ValueError:
            pass

        parsed_name = parse_name(link['data-name'])

        data['gpus'].append({
            'manufacturer': link['data-brand'],
            'name': link['data-name'],
            'price': price,
            'id': id,
            'brand': parsed_name['brand'],
            'line': parsed_name['line'].strip(),
            'model': parsed_name['model'],
            'stock': stock,
            'sku': sku,
            'link': f'https://www.microcenter.com{link2["href"]}'
        })

    return data

def parse_name(name):
    tokens = iter(name.split(' '))
    data = {}

    cur = next(tokens)

    # get brand
    data['brand'] = cur

    # build line name
    cur = next(tokens)
    line = ''
    while cur.isalpha() == True:
        line = f'{line} {cur}'
        cur = next(tokens)

    # build model name
    model = cur
    cur = next(tokens)
    if cur == 'XT' or cur == 'Ti':
        model = f'{model} {cur}'
        cur = next(tokens)

    data['line'] = line
    data['model'] = model

    return data


def main():
    parser = argparse.ArgumentParser(description='Scrape Microcenter website for GPU listings')
    parser.add_argument('-s', '--source', required=True, dest='source', action='store', help='Source URL to scrape')
    parser.add_argument('-p', '--pretty', dest='pretty', action='store_true', help='Format the result JSON for human readability')
    args = parser.parse_args()

    data = scrape(args.source)

    if args.pretty:
        print(json.dumps(data, indent=2, sort_keys=True))
    else:
        print(json.dumps(data))

if __name__ == "__main__":
    main()