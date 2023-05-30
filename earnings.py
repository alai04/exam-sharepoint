from datetime import date

from dotenv import dotenv_values
import yfinance as yf
from office365.sharepoint.client_context import ClientContext
from office365.sharepoint.listitems.caml.query import CamlQuery

config = dotenv_values()
tickers = config['TICKER_LIST'].split(',')
print(tickers)

ctx = ClientContext(config['SITE_URL'])
ctx.with_user_credentials(config['USERNAME'], config['PASSWORD'])
site = ctx.site.get().execute_query()
print(site.url)
list = ctx.web.lists.get_by_title('Earnings Calendar')


def getEarningsDate(ticker):
    try:
        tk = yf.Ticker(ticker)
        dts = tk.earnings_dates.index.tolist()
        dts = [dt for dt in [ts.date() for ts in dts] if dt >= date.today()]
        dt = min(dts)
        # print(tk.info)
        return {
            'Title': tk.info['shortName'],
            'Description': tk.info['longName'],
            'EventDate': dt.strftime("%Y-%m-%d"),
            'EndDate': dt.strftime("%Y-%m-%d"),
            'fAllDayEvent': 'True',
        }
    except:
        return None


def update1Ticker(ev):
    print(ev)
    if len(ev['Title']) == 0:
        return

    filter_text = "Title eq '{0}'".format(ev['Title'])
    items = list.items.filter(filter_text).get().execute_query()
    if len(items) == 0:
        list.add_item(ev).execute_query()
    else:
        print(items)
        items[0].validate_update_list_item(ev).execute_query()


if __name__ == '__main__':
    items = list.items.get().execute_query()
    print('There are {0} items in Earnings Calendar.'.format(len(items)))

    for ticker in tickers:
        ev = getEarningsDate(ticker)
        if ev:
            update1Ticker(ev)

    items = list.items.get().execute_query()
    print('There are {0} items in Earnings Calendar.'.format(len(items)))
