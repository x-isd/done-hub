<p align="right">
   <a href="./README.md">中文</a> | <strong>English</strong>
</p>

<p align="center">
   <picture>
   <img style="width: 80%" src="https://pic1.imgdb.cn/item/6846e33158cb8da5c83eb1eb.png" alt="image__3_-removebg-preview.png"> 
    </picture>
</p>

<div align="center">

_This project is developed based on [one-hub](https://github.com/MartialBE/one-api)_

<a href="https://t.me/+LGKwlC_xa-E5ZDk9">
  <img src="https://img.shields.io/badge/Telegram-AI Wave Community-0088cc?style=for-the-badge&logo=telegram&logoColor=white" alt="Telegram Group" />
</a>

<sup><i>AI Wave Community</i></sup> · <sup><i>(Free API & AI bots provided in the group)</i></sup>

### [📚 Original Project Documentation](https://one-hub-doc.vercel.app/)

</div>


## Current Differences from Original Version (Latest Image)

- Support for batch deletion of channels
- Support for deleting specific parameters in channel extra parameters
- Support for model variable replacement in channel BaseURL
- Support for extra parameter pass-through in native /gemini image generation requests
- Support for custom channels using Claude native routes - integration with ClaudeCode
- Support for VertexAI channels using Claude native routes - integration with ClaudeCode
- Support for configuring multiple Regions under VertexAI channels, randomly selecting a Region for each request
- Support for gemini-2.0-flash-preview-image-generation text-to-image/image-to-image, compatible with OpenAI dialogue interface
- Added user grouping functionality for batch channel addition
- Added time period conditions in recharge statistics in analysis function
- Added RPM / TPM / CPM display in analysis function
- Added configuration for whether empty responses are billable (Default: billable)
- Added invitation recharge rebate function (Optional types: fixed/percentage)
- Fixed several bugs where user quota cache and DB data inconsistency caused billing anomalies
- Fixed invitation record field missing bug
- Fixed payment callback bug in multi-instance deployment
- Fixed timezone hardcoding bug affecting statistical data
- Fixed bug caused by allowing cf cache under API routes
- Fixed login exception bug in system initialization under http environment
- Removed meaningless original price related styles in log function
- Optimized various UI interactions
- ...

## Deployment

> Follow the original deployment tutorial and replace the image with `deanxv/done-hub`.

> Database compatible, original version can directly pull this image for migration.

## Acknowledgements

- This program uses the following open source projects
  - [one-hub](https://github.com/MartialBE/one-api) as the foundation of this project

Thanks to the authors and contributors of the above projects