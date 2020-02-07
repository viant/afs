## Feb 7 2020 0.15.1
  * Added option.PreSign
  * Added option.content.Meta
  
## Feb 5 2020 0.14.2
  * Updated destURL composition on move operation
  * Updated private matcher.Basic filed

## Jan 19 2020 0.14.0
  * Added url.JoinUNC helper function
  * Added url.IsRelative helper function
  * Add check for nil reader in Copy
      
## Dec 20 2019 0.12.0
  * Added option.Proxy 

## Dec 2 2019 0.11.0
  * Added caching service 

## Nov 1 2019 0.9.0
  * Added Storager.Get(ctx context.Context, location string, options ...Option) (os.FileInfo, error) interface
  * Added Getter.Object(ctx context.Context, URL string, options ...Option) (Object, error) interface
  * Optimize Exists, Object operation (to avoid expensive list operation)
  * Added base.Storager

## Nov 1 2019 0.7.0
  * Renamed option.Checksum to SkipChecksum
    
## October 30 2019 0.6.1
  * Implemented StoragerAuthTracker on scp 
  * Added option.Auth to control auth reusibility

## October 28 2019 0.6.0
  * Added Sizer interface
  * Added Checksum option with Skip flag (upload)
  * Add Stream option with PartSize (download)
  * Added base.StreamReader
  * Signature change (from []byte to io.Reader)
     - Storager.Upload
     - Storager.Create

## October 15 2019 0.5.0
  * Added AuthTracker

## October 15 2019 0.4.1
  * Update copy implementation
  * Added url.IsSchemeEquals

## October 12 2019 0.3.2
  * Added FileCopier 
  * Streamlined internal cloud API
  
## October 9 2019 0.3.1
  * Patched URL for single file list operation
  * Streamlined object function 
  * Streamlined exists function
  
## October 4 2019 0.3.0
  * Optimized base upload
  * Added option.Recursive for the list operation
    
## October 1 2019 0.2.1

  * Patched default BatchUploader close
  * Added ssh proto

## October 1 2019 0.2.0

  * Renamed Matcher func to Match,
  * Introduced Matcher interface
  * Added option.GetListOptions helper
  * Added option.GetWalkOptions helper


## August 20 2019

  * Initial Release.

