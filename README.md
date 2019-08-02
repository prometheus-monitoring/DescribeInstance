# DescribeInstance
Công cụ tự động lấy các thông tin của server từ **amazone**, **google cloud**, hoặc từ **database mysql**. Sau đó chuẩn hóa các thông tin đó với định dạng các tập tin json.
**Cách sử dụng:**   DescribeInstance [\<flags>] 

````
Flags:
-h, --help  Show context-sensitive help (also try --help-long and --help-man).
 -c, --config.file="config.yml"  DescribeInstance configuration file path.
 -m, --add.manual                Add targets munual
 -p, --path                      Destination directory store target file
 -d, --datacenter=DATACENTER  
          Choose data center:
             - all: Get all targets from the data center include aws, gcp, vng
             - aws: Get all targets from the amazone web services
             - gcp: Get all targets from the google cloud
             - vng: Get all targets from the VN data center(If add target manual please choose vng_newfarm, vng_oldfarm or vng_singapore)
````
