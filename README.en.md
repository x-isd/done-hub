<p align="right">
   <strong>English</strong> | <a href="./README.md">中文</a>
</p>

<p align="center">
   <picture>
   <img style="width: 80%" src="https://pic1.imgdb.cn/item/6846e33158cb8da5c83eb1eb.png" alt="image__3_-removebg-preview.png"> 
    </picture>
</p>

<div align="center">

_This project is a secondary development based on [one-hub](https://github.com/MartialBE/one-api)_

<a href="https://t.me/+LGKwlC_xa-E5ZDk9">
  <img src="https://img.shields.io/badge/Telegram-AI Wave Discussion Group-0088cc?style=for-the-badge&logo=telegram&logoColor=white" alt="Telegram Discussion Group" />
</a>

<sup><i>AI Wave Community</i></sup> · <sup><i>(Free API and AI bots provided in the group)</i></sup>

### [📚 Original Project Documentation](https://one-hub-doc.vercel.app/)

</div>

## Current Differences from the Original Version

- Support for Gemini model to display thinking process
- Support for /gemini native requests on Vertex-AI channel
- Support for global region on Vertex-AI channel
- Support for image generation models on Vertex-AI channel, compatible with OpenAI image generation API
- Support for video parsing requests on Vertex-AI channel aligned with OpenAI API
- Support for additional parameter pass-through in /gemini native image generation requests
- Support for thinking parameters in /gemini native conversation requests
- Support for gemini-2.0-flash-preview-image-generation text-to-image/image-to-image, compatible with OpenAI conversation API
- Support for batch deletion of channels
- Added RPM TPM display in analytics functionality
- Added invitation recharge rebate feature (optional types: fixed/percentage)
- Added configuration for whether empty replies are billable (default: billable)
- Fixed bug where editing model price type was ineffective
- Removed meaningless original price related styles in log functionality
- ...

## Deployment

> Follow the original deployment tutorial and replace the image with `deanxv/done-hub`.

> Database compatible, original version can directly pull this image for migration.

## Acknowledgements

- This program uses the following open source projects
  - [one-hub](https://github.com/MartialBE/one-api) as the foundation for this project

Thanks to the authors and contributors of the above projects
