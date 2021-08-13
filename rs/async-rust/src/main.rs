use chrono::prelude::*;
use clap::Clap;
use std::io::{Error, ErrorKind};
use yahoo_finance_api as yahoo;

#[derive(Clap)]
#[clap(
    version = "1.0",
    author = "Claus Matzinger",
    about = "A Manning LiveProject: async Rust"
)]
struct Opts {
    #[clap(short, long, default_value = "AAPL,MSFT,UBER,GOOG")]
    symbols: String,
    #[clap(short, long)]
    from: String,
}

///
/// A trait to provide a common interface for all signal calculations.
///
trait AsyncStockSignal {

    ///
    /// The signal's data type.
    ///
    type SignalType;

    ///
    /// Calculate the signal on the provided series.
    ///
    /// # Returns
    ///
    /// The signal (using the provided type) or `None` on error/invalid data.
    ///
    fn calculate(&self, series: &[f64]) -> Option<Self::SignalType>;
}

///
/// Calculates the absolute and relative difference between the beginning and ending of an f64 series. The relative difference is relative to the beginning.
///
/// # Returns
///
/// A tuple `(absolute, relative)` difference.
///
fn price_diff(a: &[f64]) -> Option<(f64, f64)> {
    if !a.is_empty() {
        // unwrap is safe here even if first == last
        let (first, last) = (a.first().unwrap(), a.last().unwrap());
        let abs_diff = last - first;
        let first = if *first == 0.0 { 1.0 } else { *first };
        let rel_diff = abs_diff / first;
        Some((abs_diff, rel_diff))
    } else {
        None
    }
}

struct PriceDifference;

impl AsyncStockSignal for PriceDifference {
    type SignalType = (f64, f64);

    fn calculate(&self, series: &[f64]) -> Option<Self::SignalType> {
        price_diff(series)
    }
}

///
/// Find the minimum in a series of f64
///
fn min(series: &[f64]) -> Option<f64> {
    if series.is_empty() {
        None
    } else {
        Some(series.iter().fold(f64::MAX, |acc, q| acc.min(*q)))
    }
}

struct MinPrice;

impl AsyncStockSignal for MinPrice {
    type SignalType = f64;

    fn calculate(&self, series: &[f64]) -> Option<Self::SignalType> {
        min(series)
    }
}


///
/// Find the maximum in a series of f64
///
fn max(series: &[f64]) -> Option<f64> {
    if series.is_empty() {
        None
    } else {
        Some(series.iter().fold(f64::MIN, |acc, q| acc.max(*q)))
    }
}

struct MaxPrice;

impl AsyncStockSignal for MaxPrice {
    type SignalType = f64;

    fn calculate(&self, series: &[f64]) -> Option<Self::SignalType> {
        max(series)
    }
}

///
/// Window function to create a simple moving average
///
fn n_window_sma(n: usize, series: &[f64]) -> Option<Vec<f64>> {
    if !series.is_empty() && n > 1 {
        Some(
            series
                .windows(n)
                .map(|w| w.iter().sum::<f64>() / w.len() as f64)
                .collect(),
        )
    } else {
        None
    }
}

struct WindowedSMA {
    window_size: usize,
}

impl AsyncStockSignal for WindowedSMA {
    type SignalType = Vec<f64>;

    fn calculate(&self, series: &[f64]) -> Option<Self::SignalType> {
        n_window_sma(self.window_size, series)
    }
}

///
/// Retrieve data from a data source and extract the closing prices. Errors during download are mapped onto io::Errors as InvalidData.
///
fn fetch_closing_data(
    symbol: &str,
    beginning: &DateTime<Utc>,
    end: &DateTime<Utc>,
) -> std::io::Result<Vec<f64>> {
    let provider = yahoo::YahooConnector::new();

    let response = provider
        .get_quote_history(symbol, *beginning, *end)
        .map_err(|_| Error::from(ErrorKind::InvalidData))?;
    let mut quotes = response
        .quotes()
        .map_err(|_| Error::from(ErrorKind::InvalidData))?;
    if !quotes.is_empty() {
        quotes.sort_by_cached_key(|k| k.timestamp);
        Ok(quotes.iter().map(|q| q.adjclose as f64).collect())
    } else {
        Ok(vec![])
    }
}

fn main() -> std::io::Result<()> {
    let opts = Opts::parse();
    let from: DateTime<Utc> = opts.from.parse().expect("Couldn't parse 'from' date");
    let to = Utc::now();

    // a simple way to output a CSV header
    println!("period start,symbol,price,change %,min,max,30d avg");
    for symbol in opts.symbols.split(',') {
        let closes = fetch_closing_data(&symbol, &from, &to)?;
        if !closes.is_empty() {
                // min/max of the period. unwrap() because those are Option types
                let period_max: f64 = max(&closes).unwrap();
                let period_min: f64 = min(&closes).unwrap();
                let last_price = *closes.last().unwrap_or(&0.0);
                let (_, pct_change) = price_diff(&closes).unwrap_or((0.0, 0.0));
                let sma = n_window_sma(30, &closes).unwrap_or_default();

            // a simple way to output CSV data
            println!(
                "{},{},${:.2},{:.2}%,${:.2},${:.2},${:.2}",
                from.to_rfc3339(),
                symbol,
                last_price,
                pct_change * 100.0,
                period_min,
                period_max,
                sma.last().unwrap_or(&0.0)
            );
        }
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    #![allow(non_snake_case)]
    use super::*;

    #[test]
    fn test_PriceDifference_calculate() {
        let signal = PriceDifference {};
        assert_eq!(signal.calculate(&[]), None);
        assert_eq!(signal.calculate(&[1.0]), Some((0.0, 0.0)));
        assert_eq!(signal.calculate(&[1.0, 0.0]), Some((-1.0, -1.0)));
        assert_eq!(
            signal.calculate(&[2.0, 3.0, 5.0, 6.0, 1.0, 2.0, 10.0]),
            Some((8.0, 4.0))
        );
        assert_eq!(
            signal.calculate(&[0.0, 3.0, 5.0, 6.0, 1.0, 2.0, 1.0]),
            Some((1.0, 1.0))
        );
    }

    #[test]
    fn test_MinPrice_calculate() {
        let signal = MinPrice {};
        assert_eq!(signal.calculate(&[]), None);
        assert_eq!(signal.calculate(&[1.0]), Some(1.0));
        assert_eq!(signal.calculate(&[1.0, 0.0]), Some(0.0));
        assert_eq!(
            signal.calculate(&[2.0, 3.0, 5.0, 6.0, 1.0, 2.0, 10.0]),
            Some(1.0)
        );
        assert_eq!(
            signal.calculate(&[0.0, 3.0, 5.0, 6.0, 1.0, 2.0, 1.0]),
            Some(0.0)
        );
    }

    #[test]
    fn test_MaxPrice_calculate() {
        let signal = MaxPrice {};
        assert_eq!(signal.calculate(&[]), None);
        assert_eq!(signal.calculate(&[1.0]), Some(1.0));
        assert_eq!(signal.calculate(&[1.0, 0.0]), Some(1.0));
        assert_eq!(
            signal.calculate(&[2.0, 3.0, 5.0, 6.0, 1.0, 2.0, 10.0]),
            Some(10.0)
        );
        assert_eq!(
            signal.calculate(&[0.0, 3.0, 5.0, 6.0, 1.0, 2.0, 1.0]),
            Some(6.0)
        );
    }

    #[test]
    fn test_WindowedSMA_calculate() {
        let series = vec![2.0, 4.5, 5.3, 6.5, 4.7];

        let signal = WindowedSMA { window_size: 3 };
        assert_eq!(
            signal.calculate(&series),
            Some(vec![3.9333333333333336, 5.433333333333334, 5.5])
        );

        let signal = WindowedSMA { window_size: 5 };
        assert_eq!(signal.calculate(&series), Some(vec![4.6]));

        let signal = WindowedSMA { window_size: 10 };
        assert_eq!(signal.calculate(&series), Some(vec![]));
    }
}
