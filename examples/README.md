**GeoNet FITS Data Access and Plotting Jupyter Notebooks**

 

This series of five Jupyter Notebooks covers the discovery, acquisition, and plotting of data from the GeoNet FITS (FIeld Time Series) database using Python. Each Notebook contains both documentation and code, and all code is unlicensed and free to use. A summary of each Notebook is provided below. 

 

Notebook 1 is a brief introduction to FITS data access and plotting using Python and the pandas module. It covers:

- Building a FITS query
- Querying data from FITS
- Viewing FITS data in Python
- Plotting FITS data in Python
- Saving FITS data as a csv

 

Notebook 2 is an expansion of the first Notebook and covers how to access and plot FITS data for multiple sites and/or data types. It covers:

- Building FITS queries for multiple sites and/or data types
- Using functions to handle FITS data querying
- Plotting multiple FITS datasets on the same plot
- Using subplotting to plot multiple FITS datasets on the same figure
- An example of data manipulation using pandas

 

Notebook 3 focuses on data discovery for one or more sites. It covers:

- Discovery of what data types exist for a site
- Plotting all data for up to 9 data types at a site
- Saving plots as PNG images

 

Notebook 4 is an example of making a simple plot of site positions in FITS and displaying it in Python. It covers:

- How plotting extents are defined in FITS
- How to build a map query in FITS
- Plotting SVG images in Python

 

Notebook 5 is for advanced users, as the code is much more complex and the purpose very focused. This notebook concentrates on discovering which sites have data for one or more data types or data collection methods. It covers:

- Discovery of which sites have data for one or more data types or data collection methods
- Plotting all data (or a window thereof) of the data types / collection methods specified for many sites
